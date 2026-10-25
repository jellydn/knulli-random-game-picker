"""Exercise the real loopback transport, including ambiguous launch outcomes."""
import http.server
import os
import subprocess
import sys
import threading
import time

requests = []


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
        requests.append(('GET', self.path, b'', self.headers['Host']))
        if self.path in responses:
            self.connection.sendall(responses[self.path])
            self.close_connection = True
            return
        if self.path == '/missing':
            self.send_response(404)
            self.send_header('Content-Length', '0')
            self.end_headers()
        else:
            self.send_response(200)
            self.send_header('Content-Length', '2')
            self.end_headers()
            self.wfile.write(b'[]')

    def do_POST(self):
        data = self.rfile.read(int(self.headers['Content-Length']))
        requests.append(('POST', self.path, data, self.headers['Content-Type']))
        if self.path == '/timeout':
            # Accept the launch but never acknowledge it. The transport's own
            # deadline, not subprocess's watchdog, must stop the request.
            release.wait(7)
            return
        if self.path in responses:
            self.connection.sendall(responses[self.path])
            self.close_connection = True
            return
        self.send_response(200)
        # A truncated success response must not cause the client to re-submit.
        self.send_header('Content-Length', '2' if self.path == '/truncated' else '0')
        self.end_headers()


responses = {
    '/chunked': b'HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n2\r\n[]\r\n0\r\n\r\n',
    '/unframed': b'HTTP/1.1 200 OK\r\nConnection: close\r\n\r\n[]',
    '/oversized': b'HTTP/1.1 200 OK\r\nContent-Length: 67108865\r\n\r\n',
    '/conflicting-length': b'HTTP/1.1 200 OK\r\nContent-Length: 0\r\nContent-Length: 2\r\n\r\n[]',
    '/bad-length': b'HTTP/1.1 200 OK\r\nContent-Length: -1\r\n\r\n',
    '/redirect': b'HTTP/1.1 307 Temporary Redirect\r\nLocation: http://127.0.0.1:1234/launch\r\nContent-Length: 0\r\n\r\n',
    '/disconnect': b'',
}
release = threading.Event()
server = http.server.ThreadingHTTPServer(('127.0.0.1', 1234), Handler)
thread = threading.Thread(target=server.serve_forever, daemon=True)
thread.start()
try:
    def run(method, path, body=''):
        # A deliberately unusable proxy must never affect fixed-loopback IPC.
        env = dict(os.environ, HTTP_PROXY='http://127.0.0.1:1',
                   HTTPS_PROXY='http://127.0.0.1:1', ALL_PROXY='http://127.0.0.1:1',
                   NO_PROXY='', no_proxy='')
        return subprocess.run([sys.argv[1], method, path, body], capture_output=True,
                              text=True, timeout=8, env=env)

    response = run('GET', '/systems')
    assert response.returncode == 0 and response.stdout == '[]', response
    assert run('GET', '/missing').returncode == 1
    escaped = '/systems/a%2Fb%20c%25/games'
    assert run('GET', escaped).returncode == 0
    assert requests[-1][1] == escaped, requests[-1]
    assert all(request[3] == '127.0.0.1:1234' for request in requests), requests
    rom = "/test/日本語/a 'quoted'; $(no-shell) %2F\n game.rom"
    assert run('POST', '/launch', rom).returncode == 0
    assert run('POST', '/truncated', rom).returncode == 1
    for route in responses:
        before = len(requests)
        result = run('POST', route, rom)
        assert result.returncode == 1, (route, result)
        assert len(requests) == before + 1, (route, requests[before:])
    before = len(requests)
    started = time.monotonic()
    result = run('POST', '/timeout', rom)
    elapsed = time.monotonic() - started
    assert result.returncode == 1 and 4 <= elapsed < 7, (result, elapsed)
    assert len(requests) == before + 1, requests[before:]
    posts = [request for request in requests if request[0] == 'POST']
    expected_routes = ['/launch', '/truncated', *responses, '/timeout']
    assert posts == [('POST', route, rom.encode(), 'text/plain')
                     for route in expected_routes], posts
    print('PASS: exact route/body bytes, fixed loopback ignoring proxies, bounded framing, '
          'redirect rejection, disconnect/truncation/timeout and one POST without retries')
finally:
    release.set()
    server.shutdown()
    server.server_close()
    thread.join()
