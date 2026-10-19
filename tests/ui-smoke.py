"""Run the actual SDL UI against a local ES fixture. Never starts an emulator."""
import binascii
import html
import http.server
import json
import os
from pathlib import Path
import struct
import subprocess
import sys
import tempfile
import threading
import time
import zlib

binary = Path(sys.argv[1]).resolve()
captures = Path(sys.argv[2]).resolve()
captures.mkdir(parents=True, exist_ok=True)
requests = []
reads = []


def png(path):
    def chunk(kind, data):
        return struct.pack('>I', len(data)) + kind + data + struct.pack('>I', binascii.crc32(kind + data))
    pixels = b''.join(b'\0' + bytes([36, 130 + y // 3, 115]) * 160 for y in range(200))
    path.write_bytes(b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', struct.pack('>IIBBBBB', 160, 200, 8, 2, 0, 0, 0)) +
                     chunk(b'IDAT', zlib.compress(pixels)) + chunk(b'IEND', b''))


with tempfile.TemporaryDirectory(prefix='picker-ui-') as directory:
    root = Path(directory)
    config = root / 'config'
    roms = root / 'relocated games'
    config.mkdir()
    roms.mkdir()
    rom = roms / "fixture 'quoted'; $.rom"
    rom.write_text('synthetic ROM fixture, not executable')
    png(roms / 'art.png')
    (config / 'es_systems_custom.cfg').write_text('<systemList><system><name>fixture</name><fullname>Fixture Console</fullname><path>' + html.escape(str(roms)) + '</path></system></systemList>')
    (roms / 'gamelist.xml').write_text('<gameList><game><path>./' + html.escape(rom.name) + '</path><image>./art.png</image></game></gameList>')

    class Handler(http.server.BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def do_GET(self):
            reads.append(self.path)
            if self.path == '/systems':
                data = [{'name': 'fixture', 'gamesystem': 'true', 'collection': 'false'}]
            elif self.path == '/systems/fixture/games':
                data = [{'path': str(rom), 'name': 'A synthetic game for testing', 'favorite': 'true'}]
            else:
                self.send_error(404)
                return
            body = json.dumps(data).encode()
            self.send_response(200)
            self.send_header('Content-Length', str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def do_POST(self):
            body = self.rfile.read(int(self.headers['Content-Length']))
            # The real window must be gone before the launch request is sent.
            window = subprocess.run(['xdotool', 'search', '--name', 'KNULLI Random Game Picker'], env=environment,
                                    capture_output=True)
            requests.append((self.path, body, window.returncode))
            self.send_response(200)
            self.send_header('Content-Length', '0')
            self.end_headers()

    server = http.server.HTTPServer(('127.0.0.1', 1234), Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    environment = dict(os.environ)
    read_fd, write_fd = os.pipe()
    display = subprocess.Popen(['Xvfb', '-displayfd', str(write_fd), '-screen', '0', '1280x720x24', '-nolisten', 'tcp'],
                               pass_fds=(write_fd,), stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    os.close(write_fd)
    with os.fdopen(read_fd) as pipe:
        environment['DISPLAY'] = ':' + pipe.readline().strip()
    environment['PICKER_STATE_DIR'] = str(root / 'state')
    process = None
    try:
        # KNULLI's normal ES->SDL mapping swaps both face-button pairs.
        # Feed SDL events, not keyboard equivalents, to catch the reported bug.
        fifo = root / 'controller-input'
        os.mkfifo(fifo)
        with os.fdopen(os.open(fifo, os.O_RDWR | os.O_NONBLOCK), 'wb', buffering=0) as buttons:
            for layout, a, b, x, y in [('knulli', 1, 0, 3, 2), ('xbox', 0, 1, 2, 3)]:
                controller_environment = dict(environment, PICKER_CONTROLLER_LAYOUT=layout,
                                              PICKER_TEST_INPUT_FIFO=str(fifo),
                                              LD_PRELOAD=str(binary.parent / 'libcontroller-events.so'))
                before_launch, before_reads = len(requests), len(reads)
                process = subprocess.Popen([str(binary), '--windowed', '--size', '1280x720', '--es-config', str(config)], env=controller_environment)
                for _ in range(100):
                    if len(reads) >= before_reads + 2:
                        break
                    assert process.poll() is None
                    time.sleep(0.05)
                time.sleep(0.5)
                def pad(button):
                    buttons.write(bytes([button]))
                    time.sleep(0.2)
                pad(a)  # ES A selects; it must not exit or launch.
                assert process.poll() is None, f'{layout}: A exited instead of selecting'
                assert len(requests) == before_launch
                pad(x)  # ES X rerolls without requesting a rescan.
                assert len(reads) == before_reads + 2, f'{layout}: X rescanned instead of rerolling'
                pad(b)  # ES B returns to filters without launching.
                assert process.poll() is None and len(requests) == before_launch, f'{layout}: B did not return to filters'
                pad(y)  # ES Y must rescan, not reroll or do nothing.
                for _ in range(100):
                    if len(reads) >= before_reads + 4:
                        break
                    time.sleep(0.05)
                assert len(reads) == before_reads + 4, f'{layout}: Y did not rescan'
                time.sleep(0.3)
                pad(a)
                pad(a)
                assert process.wait(timeout=10) == 0
                assert len(requests) == before_launch + 1
                assert requests[-1] == ('/launch', str(rom).encode(), 1)
                print(f'PASS: {layout} controller events: A pick/play, B back, X reroll, Y rescan')
        requests.clear()
        for size in ('1280x720', '640x480'):
            image = roms / 'art.png'
            if size == '640x480':
                image = roms / 'art.jpg'
                subprocess.run(['convert', str(roms / 'art.png'), str(image)], check=True)
                gamelist = roms / 'gamelist.xml'
                gamelist.write_text(gamelist.read_text().replace('art.png', 'art.jpg'))
            # Each width starts from the same filter state, not the prior run's saved settings.
            (root / 'state/settings.json').unlink(missing_ok=True)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = ''
            for _ in range(100):
                search = subprocess.run(['xdotool', 'search', '--name', 'KNULLI Random Game Picker'], env=environment, capture_output=True, text=True)
                if search.returncode == 0:
                    window = search.stdout.splitlines()[0]
                    break
                assert process.poll() is None, 'UI exited before opening'
                time.sleep(0.05)
            assert window, 'UI did not open'
            def key(name):
                subprocess.run(['xdotool', 'key', '--window', window, name], env=environment, check=True, stderr=subprocess.DEVNULL)
                time.sleep(0.1)
            def capture(state):
                path = captures / f'{state}-{size}.png'
                for _ in range(100):
                    subprocess.run(['import', '-display', environment['DISPLAY'], '-window', window, str(path)], check=True)
                    colors = subprocess.check_output(['identify', '-format', '%k', str(path)], text=True)
                    if int(colors) > 16:
                        return
                    assert process.poll() is None, 'UI exited before rendering'
                    time.sleep(0.05)
                raise AssertionError('UI capture remained blank')
            time.sleep(0.3)
            capture('filters')
            key('Return')
            capture('preview-art')
            original_art = image.read_bytes()
            image.write_bytes(b'not a valid image')
            key('x')
            capture('fallback')
            image.write_bytes(original_art)
            for _ in range(100):
                key('x')
            assert process.poll() is None, 'UI crashed during rerolls'
            key('Escape')
            key('Down')
            key('Return')  # favorites toggle
            capture('favorites')
            key('Down')
            key('Return')  # exclude recent
            key('Up')
            key('Up')
            key('Return')
            capture('empty-recent')
            key('Down')
            key('Down')
            key('Down')
            key('Return')  # clear history
            key('Up')
            key('Up')
            key('Up')
            key('Return')  # select again
            key('Return')  # real single-submit handoff to fake ES
            assert process.wait(timeout=10) == 0
            assert requests[-1] == ('/launch', str(rom).encode(), 1), requests
            assert len(requests) == (1 if size == '1280x720' else 2), requests
            print(f'PASS: {size}, scraped-art preview, 100 rerolls, filters, empty pool, SDL teardown before one launch')
    finally:
        if process and process.poll() is None:
            process.terminate()
            process.wait(timeout=8)
        display.terminate()
        display.wait(timeout=5)
        server.shutdown()
        server.server_close()
        thread.join()
