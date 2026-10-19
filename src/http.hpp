#pragma once
#include <string>

namespace picker {
// Local EmulationStation IPC only. No redirects, DNS, proxies, or automatic retries.
std::string es_request(const std::string& method, const std::string& route,
                       const std::string& body = "");
}
