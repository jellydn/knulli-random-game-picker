"""Run the actual SDL UI against a local ES fixture. Never starts an emulator."""
import binascii
import html
import http.server
import json
import os
from pathlib import Path
import signal
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
other_active = False
library_available = True
api_unavailable = False
fixture_title = 'A synthetic game for testing'
scan_allowed = threading.Event()
scan_allowed.set()


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
    other = root / 'other games'
    other.mkdir()
    other_rom = other / 'not-a-favorite.rom'
    other_rom.write_text('another synthetic ROM fixture, not executable')
    (config / 'es_systems_custom.cfg').write_text('<systemList><system><name>fixture</name><fullname>Fixture Console</fullname><path>' + html.escape(str(roms)) + '</path></system><system><name>other</name><fullname>Other Console</fullname><path>' + html.escape(str(other)) + '</path></system></systemList>')
    (roms / 'gamelist.xml').write_text('<gameList><game><path>./' + html.escape(rom.name) + '</path><image>./art.png</image></game></gameList>')
    (other / 'gamelist.xml').write_text('<gameList><game><path>./not-a-favorite.rom</path></game></gameList>')

    class Handler(http.server.BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def do_GET(self):
            reads.append(self.path)
            if self.path == '/systems':
                scan_allowed.wait(timeout=5)
                if api_unavailable:
                    self.send_error(503)
                    return
                data = [{'name': 'fixture', 'gamesystem': 'true', 'collection': 'false'}] if library_available else []
                if other_active and library_available:
                    data.append({'name': 'other', 'gamesystem': 'true', 'collection': 'false'})
            elif self.path == '/systems/fixture/games':
                data = [{'path': str(rom), 'name': fixture_title, 'favorite': 'true'}]
            elif self.path == '/systems/other/games':
                data = [{'path': str(other_rom), 'name': 'Saved system game', 'favorite': 'false'}]
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
    recording = None

    def wait_window():
        for _ in range(100):
            search = subprocess.run(['xdotool', 'search', '--name', 'KNULLI Random Game Picker'], env=environment, capture_output=True, text=True)
            if search.returncode == 0:
                return search.stdout.splitlines()[0]
            assert process.poll() is None, 'UI exited before opening'
            time.sleep(0.05)
        raise AssertionError('UI did not open')

    def wait_reads(count):
        for _ in range(100):
            if len(reads) >= count:
                time.sleep(0.3)
                return
            assert process.poll() is None, 'UI exited while reading the library'
            time.sleep(0.05)
        raise AssertionError('Library scan did not finish')

    def key(window, name):
        # Use focused XTest events, not XSendEvent against a window which may
        # close on key-down before xdotool can deliver key-up.
        subprocess.run(['xdotool', 'windowfocus', '--sync', window, 'key', name], env=environment, check=True)
        time.sleep(0.1)

    def capture(window, state, size):
        path = captures / f'{state}-{size}.png'
        for _ in range(100):
            subprocess.run(['import', '-display', environment['DISPLAY'], '-window', window, str(path)], check=True)
            colors = subprocess.check_output(['identify', '-format', '%k', str(path)], text=True)
            if int(colors) > 16:
                return
            assert process.poll() is None, 'UI exited before rendering'
            time.sleep(0.05)
        raise AssertionError('UI capture remained blank')

    try:
        unknown = subprocess.run([str(binary), '--firmware', 'unverified'], env=environment, capture_output=True)
        assert unknown.returncode == 1 and b'KNULLI only' in unknown.stderr
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
                window = wait_window()
                wait_reads(before_reads + 2)
                def pad(button):
                    buttons.write(bytes([button]))
                    time.sleep(0.2)
                capture(window, f'{layout}-startup', '1280x720')
                assert len(requests) == before_launch, 'Startup must not launch automatically'
                pad(a)  # The first A after startup plays the automatically selected game.
                assert process.wait(timeout=5) == 0, f'{layout}: startup requires another selection step'
                assert len(requests) == before_launch + 1
                assert requests[-1] == ('/launch', str(rom).encode(), 1)

                before_launch, before_reads = len(requests), len(reads)
                process = subprocess.Popen([str(binary), '--windowed', '--size', '1280x720', '--es-config', str(config)], env=controller_environment)
                wait_window()
                wait_reads(before_reads + 2)
                pad(x)  # ES X rerolls without requesting a rescan.
                assert len(reads) == before_reads + 2, f'{layout}: X rescanned instead of rerolling'
                pad(y)  # ES Y opens settings, without a scan or launch.
                assert len(reads) == before_reads + 2 and len(requests) == before_launch, f'{layout}: Y did not open settings'
                pad(y)  # Y inside settings retains the existing rescan action.
                wait_reads(before_reads + 4)
                assert len(reads) == before_reads + 4, f'{layout}: settings rescan failed'
                pad(a)  # Settings System row selects a game as before.
                pad(b)  # ES B returns to filters without launching.
                assert process.poll() is None and len(requests) == before_launch, f'{layout}: B did not return to filters'
                pad(b)  # B from filters still quits.
                assert process.wait(timeout=5) == 0 and len(requests) == before_launch

                before_reads = len(reads)
                process = subprocess.Popen([str(binary), '--windowed', '--size', '1280x720', '--es-config', str(config)], env=controller_environment)
                wait_window()
                wait_reads(before_reads + 2)
                pad(6)  # SDL Start remains a global quit, including the new startup view.
                assert process.wait(timeout=5) == 0 and len(requests) == before_launch
                print(f'PASS: {layout}: startup A plays once, Y settings/rescan, X reroll, B back/quit, Start quit')
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
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 2)
            capture(window, 'selector', size)
            original_art = image.read_bytes()
            image.write_bytes(b'not a valid image')
            key(window, 'x')
            capture(window, 'fallback', size)
            image.write_bytes(original_art)
            for _ in range(100):
                key(window, 'x')
            assert process.poll() is None, 'UI crashed during rerolls'
            if size == '640x480' and os.getenv('PICKER_UI_RECORD'):
                recording = subprocess.Popen(['ffmpeg', '-y', '-loglevel', 'error', '-f', 'x11grab',
                    '-video_size', size, '-framerate', '20', '-i', environment['DISPLAY'] + '.0',
                    '-c:v', 'libvpx-vp9', '-crf', '35', '-b:v', '0', os.environ['PICKER_UI_RECORD']], stdin=subprocess.PIPE)
                time.sleep(0.4)
                for pointer_x in range(55, 575, 20):
                    subprocess.run(['xdotool', 'mousemove', '--sync', str(pointer_x), '451'], env=environment, check=True)
                    time.sleep(0.04)
                time.sleep(0.8)
            key(window, 'y')
            capture(window, 'settings', size)
            key(window, 'Down')
            key(window, 'Return')  # favorites toggle
            capture(window, 'favorites', size)
            key(window, 'Down')
            key(window, 'Return')  # exclude recent
            if recording:
                time.sleep(0.8)
                recording.communicate(input=b'q', timeout=10)
                assert recording.returncode == 0
                recording = None
            key(window, 'Up')
            key(window, 'Up')
            key(window, 'Return')
            capture(window, 'empty-recent', size)
            key(window, 'Down')
            key(window, 'Down')
            key(window, 'Down')
            key(window, 'Return')  # clear history
            key(window, 'Up')
            key(window, 'Up')
            key(window, 'Up')
            key(window, 'Return')  # select again
            key(window, 'Return')  # real single-submit handoff to fake ES
            assert process.wait(timeout=10) == 0
            assert requests[-1] == ('/launch', str(rom).encode(), 1), requests
            assert len(requests) == (1 if size == '1280x720' else 2), requests
            print(f'PASS: {size}, scraped-art preview, 100 rerolls, filters, empty pool, SDL teardown before one launch')

        # A game disappearing between preview and Play must never be submitted.
        # A long title must still leave platform/status/actions readable.
        fixture_title = 'The remarkably long adventures of a synthetic hero across a relocated library with more words than fit on a single line'
        for size in ('1280x720', '640x480'):
            requests.clear()
            (root / 'state/settings.json').unlink(missing_ok=True)
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 2)
            capture(window, 'long-title', size)
            library_available = False
            key(window, 'Return')
            wait_reads(before_reads + 3)
            capture(window, 'stale-selection', size)
            assert process.poll() is None and not requests, 'Stale selection bypassed revalidation'
            key(window, 'F10')
            assert process.wait(timeout=5) == 0 and not requests
            library_available = True
            before_reads = len(reads)
            saved_bytes = (root / 'state/settings.json').read_bytes()
            process = subprocess.Popen([str(binary), '--demo', '--windowed', '--size', size], env=environment)
            window = wait_window()
            time.sleep(0.4)
            key(window, 'Return')
            capture(window, 'demo-no-launch', size)
            key(window, 'F10')
            assert process.wait(timeout=5) == 0 and not requests and len(reads) == before_reads
            assert (root / 'state/settings.json').read_bytes() == saved_bytes
            print(f'PASS: {size}: stale selection blocked; demo does not scan, launch or save')
        fixture_title = 'A synthetic game for testing'

        # The retained system contains only a nonfavorite; the other system has
        # a favorite. Ignoring either saved filter would make A launch a game.
        other_active = True
        for size in ('1280x720', '640x480'):
            requests.clear()
            saved = {'system': 'other', 'favorites': True}
            (root / 'state/settings.json').write_text(json.dumps(saved))
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 3)
            capture(window, 'empty-saved-filter', size)
            key(window, 'Return')
            assert process.poll() is None and not requests, 'Startup ignored saved system or favorites'
            key(window, 'y')
            capture(window, 'saved-settings', size)
            key(window, 'Escape')
            assert process.wait(timeout=5) == 0
            assert json.loads((root / 'state/settings.json').read_text()) == saved

            saved['favorites'] = False
            (root / 'state/settings.json').write_text(json.dumps(saved))
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 3)
            key(window, 'Return')
            assert process.wait(timeout=5) == 0
            assert requests == [('/launch', str(other_rom).encode(), 1)]
            assert json.loads((root / 'state/settings.json').read_text()) == saved

            library_available = False
            requests.clear()
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 1)
            capture(window, 'empty-library', size)
            key(window, 'Return')
            key(window, 'x')
            assert process.poll() is None and not requests, 'Empty library launched or crashed'
            key(window, 'y')
            key(window, 'y')  # Rescan remains reachable even without a selected game.
            wait_reads(before_reads + 2)
            key(window, 'Escape')
            assert process.wait(timeout=5) == 0 and not requests
            library_available = True
            print(f'PASS: {size}: saved filters applied on startup and preserved; empty selector recovers through settings')

            # Opening settings before discovery completes must not start a second
            # scan or let the background result pull the user back to the selector.
            scan_allowed.clear()
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 1)
            capture(window, 'loading', size)
            key(window, 'y')
            capture(window, 'loading-settings', size)
            assert len(reads) == before_reads + 1
            scan_allowed.set()
            wait_reads(before_reads + 3)
            key(window, 'Escape')
            try:
                assert process.wait(timeout=5) == 0 and not requests, 'Discovery stole settings focus'
            except subprocess.TimeoutExpired:
                capture(window, 'delayed-exit-timeout', size)
                process.send_signal(signal.SIGQUIT)  # Go stack dump for a stuck UI.
                process.wait(timeout=5)
                raise

            api_unavailable = True
            before_reads = len(reads)
            process = subprocess.Popen([str(binary), '--windowed', '--size', size, '--es-config', str(config)], env=environment)
            window = wait_window()
            wait_reads(before_reads + 1)
            capture(window, 'unavailable', size)
            key(window, 'Return')
            key(window, 'x')
            assert process.poll() is None and not requests, 'Failed discovery launched or crashed'
            key(window, 'y')
            api_unavailable = False
            key(window, 'y')
            wait_reads(before_reads + 4)
            key(window, 'Return')
            capture(window, 'recovered', size)
            key(window, 'Escape')
            key(window, 'Escape')
            assert process.wait(timeout=5) == 0 and not requests
            print(f'PASS: {size}: settings remains open during discovery; unavailable library recovers by rescan')
    finally:
        scan_allowed.set()
        if recording:
            recording.communicate(input=b'q', timeout=10)
        if process and process.poll() is None:
            process.terminate()
            process.wait(timeout=8)
        display.terminate()
        display.wait(timeout=5)
        server.shutdown()
        server.server_close()
        thread.join()
