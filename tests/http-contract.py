"""Exercise the real loopback transport, including ambiguous launch outcomes."""
import http.server
import subprocess
import sys
import threading

requests = []


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
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
        requests.append((self.path, data, self.headers['Content-Type']))
        self.send_response(200)
        # A truncated success response must not cause the client to re-submit.
        self.send_header('Content-Length', '2' if self.path == '/truncated' else '0')
        self.end_headers()


server = http.server.HTTPServer(('127.0.0.1', 1234), Handler)
thread = threading.Thread(target=server.serve_forever, daemon=True)
thread.start()
try:
    def run(method, path, body=''):
        return subprocess.run([sys.argv[1], method, path, body], capture_output=True, text=True, timeout=8)

    response = run('GET', '/systems')
    assert response.returncode == 0 and response.stdout == '[]', response
    assert run('GET', '/missing').returncode == 1
    rom = "/test/a 'quoted'; $(no-shell) game.rom"
    assert run('POST', '/launch', rom).returncode == 0
    assert run('POST', '/truncated', rom).returncode == 1
    assert requests == [('/launch', rom.encode(), 'text/plain'), ('/truncated', rom.encode(), 'text/plain')], requests
    print('PASS: real HTTP body, status validation, truncated response, no automatic launch retry')
finally:
    server.shutdown()
    server.server_close()
    thread.join()
