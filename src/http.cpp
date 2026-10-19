#include "http.hpp"
#include <arpa/inet.h>
#include <algorithm>
#include <cerrno>
#include <chrono>
#include <cctype>
#include <fcntl.h>
#include <poll.h>
#include <sstream>
#include <stdexcept>
#include <sys/socket.h>
#include <unistd.h>

namespace picker {
std::string es_request(const std::string& method, const std::string& route, const std::string& body) {
    if ((method != "GET" && method != "POST") || route.empty() || route[0] != '/' ||
        route.find_first_of("\r\n") != std::string::npos || body.find('\0') != std::string::npos)
        throw std::runtime_error("Invalid ES IPC request.");
    struct Socket {
        int fd = socket(AF_INET, SOCK_STREAM | SOCK_CLOEXEC | SOCK_NONBLOCK, 0);
        ~Socket() { if (fd >= 0) close(fd); }
    } socket;
    if (socket.fd < 0) throw std::runtime_error("Cannot open local ES socket.");
    const auto deadline = std::chrono::steady_clock::now() + std::chrono::seconds(5);
    auto wait = [&](short events) {
        for (;;) {
            const auto left = std::chrono::duration_cast<std::chrono::milliseconds>(deadline - std::chrono::steady_clock::now()).count();
            if (left <= 0) throw std::runtime_error("ES IPC timed out. A submitted launch may have been accepted; do not retry blindly.");
            pollfd descriptor{socket.fd, events, 0};
            const int result = poll(&descriptor, 1, static_cast<int>(left));
            if (result < 0 && errno == EINTR) continue;
            if (result <= 0) throw std::runtime_error("ES IPC unavailable or timed out.");
            return;
        }
    };
    sockaddr_in address{};
    address.sin_family = AF_INET;
    address.sin_port = htons(1234);
    address.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    if (connect(socket.fd, reinterpret_cast<sockaddr*>(&address), sizeof(address)) < 0) {
        if (errno != EINPROGRESS) throw std::runtime_error("ES local API is unavailable (127.0.0.1:1234).");
        wait(POLLOUT);
        int error = 0;
        socklen_t size = sizeof(error);
        if (getsockopt(socket.fd, SOL_SOCKET, SO_ERROR, &error, &size) < 0 || error)
            throw std::runtime_error("ES local API is unavailable (127.0.0.1:1234).");
    }
    const auto request = method + " " + route + " HTTP/1.1\r\nHost: 127.0.0.1:1234\r\nConnection: close\r\nContent-Type: text/plain\r\nContent-Length: " + std::to_string(body.size()) + "\r\n\r\n" + body;
    for (std::size_t offset = 0; offset < request.size();) {
        wait(POLLOUT);
        const auto sent = send(socket.fd, request.data() + offset, request.size() - offset, MSG_NOSIGNAL);
        if (sent < 0 && (errno == EINTR || errno == EAGAIN)) continue;
        if (sent <= 0) throw std::runtime_error("ES request interrupted. Launch outcome may be unknown.");
        offset += static_cast<std::size_t>(sent);
    }
    std::string response;
    char buffer[16384];
    for (;;) {
        wait(POLLIN);
        const auto length = recv(socket.fd, buffer, sizeof(buffer), 0);
        if (length < 0 && (errno == EINTR || errno == EAGAIN)) continue;
        if (length < 0) throw std::runtime_error("ES response interrupted. Launch outcome may be unknown.");
        if (length == 0) break;
        response.append(buffer, static_cast<std::size_t>(length));
        if (response.size() > 64 * 1024 * 1024) throw std::runtime_error("ES response exceeds 64 MiB limit.");
    }
    const auto split = response.find("\r\n\r\n");
    if (split == std::string::npos || split > 65536) throw std::runtime_error("Invalid ES HTTP headers.");
    std::istringstream headers(response.substr(0, split));
    std::string protocol;
    int status = 0;
    headers >> protocol >> status;
    if (protocol.rfind("HTTP/1.", 0) != 0 || status < 200 || status >= 300)
        throw std::runtime_error("ES returned HTTP " + std::to_string(status));
    std::string line;
    std::size_t expected = 0;
    bool has_length = false;
    while (std::getline(headers, line)) {
        std::transform(line.begin(), line.end(), line.begin(), [](unsigned char ch) { return std::tolower(ch); });
        if (line.rfind("transfer-encoding:", 0) == 0) throw std::runtime_error("Unsupported ES transfer encoding.");
        if (line.rfind("content-length:", 0) == 0) {
            expected = std::stoull(line.substr(15));
            has_length = true;
        }
    }
    const auto result = response.substr(split + 4);
    if (!has_length || result.size() != expected) throw std::runtime_error("Incomplete ES response. Launch outcome may be unknown.");
    return result;
}
}
