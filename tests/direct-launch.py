#!/usr/bin/env python3
"""Exercise the real no-UI CLI and packaged shortcut against a fake ES API."""
import html
import http.server
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import threading

binary = Path(sys.argv[1]).resolve()
root_repo = Path(__file__).resolve().parents[1]
posts = []
scans = 0
mode = "normal"

with tempfile.TemporaryDirectory(prefix="picker-direct-") as temporary:
    root = Path(temporary)
    package = root / "package with spaces"
    package.mkdir()
    shutil.copy2(binary, package / "random-game-picker")
    binary = package / "random-game-picker"
    for name in ("RandomGamePicker.sh", "Play Random Game.sh"):
        shutil.copy2(root_repo / "packaging" / name, package / name)
    config = root / "config"
    config.mkdir()
    games = {}
    for system in ("one", "two", "ports", "tools", "emulators"):
        directory = root / system
        directory.mkdir()
        games[system] = []
        for name, favorite in (("favorite 'quoted'; $.rom", True), ("ordinary.rom", False)):
            path = directory / name
            path.write_text("not executable, not a real ROM")
            games[system].append(dict(path=str(path), name=name, favorite=favorite))
    (config / "es_systems.cfg").write_text("<systemList>" + "".join(
        f"<system><name>{system}</name><path>{html.escape(str(root / system))}</path></system>"
        for system in games) + "</systemList>")

    class Handler(http.server.BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def do_GET(self):
            global scans
            if self.path == "/systems":
                scans += 1
                if mode == "scan-error" or (mode == "rescan-error" and scans == 2):
                    self.send_error(503)
                    return
                body = [dict(name=system, gamesystem=True, collection=False) for system in games]
                body.append(dict(name="collection", gamesystem=True, collection=True))
            else:
                system = self.path.split("/")[2]
                body = [dict(game) for game in games[system]]
                if mode == "empty" or (mode == "stale" and scans == 2):
                    body = []
                if mode == "hidden":
                    for game in body:
                        game["hidden"] = True
                if mode == "unfavorite" and scans == 2:
                    for game in body:
                        game["favorite"] = False
                if mode == "ambiguous" and system == "tools":
                    body += games["one"]
                if mode == "missing":
                    for game in body:
                        game["path"] += ".missing"
                if mode == "moved-system" and scans == 2:
                    body = [] if system == "one" else body + (games["one"] if system == "two" else [])
                if mode in ("own-launcher", "own-shortcut"):
                    name = "RandomGamePicker.sh" if mode == "own-launcher" else "Play Random Game.sh"
                    body = [dict(path=str(package / name), favorite=True)] if system == "one" else []
            data = json.dumps(body).encode()
            self.send_response(200)
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

        def do_POST(self):
            posts.append((self.path, self.rfile.read(int(self.headers["Content-Length"]))))
            if mode == "disconnect":
                self.close_connection = True
                return
            self.send_response(500 if mode == "launch-error" else 200)
            self.send_header("Content-Length", "0")
            self.end_headers()

    server = http.server.HTTPServer(("127.0.0.1", 1234), Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    state = root / "state"
    state.mkdir()
    settings = state / "settings.json"
    environment = dict(os.environ, PICKER_STATE_DIR=str(state), DISPLAY="",
                       SDL_VIDEODRIVER="nonexistent-no-ui-allowed", HOME=str(root))
    count = 0

    def run(saved, *, scenario="normal", launch=True, save_block=False, args=()):
        global mode, scans, count
        mode, scans = scenario, 0
        posts.clear()
        if settings.exists():
            settings.unlink()
        if saved is not None:
            settings.write_text(saved if isinstance(saved, str) else json.dumps(saved))
        before = settings.read_bytes() if settings.exists() else None
        blocker = state / "settings.json.tmp"
        if save_block:
            blocker.mkdir()
        try:
            result = subprocess.run([str(binary), "--play-random", "--es-config", str(config), *args],
                                    env=environment, capture_output=True, timeout=20)
        finally:
            if save_block:
                blocker.rmdir()
        assert result.returncode == (0 if launch and scenario == "normal" else 1), result
        assert len(posts) == int(launch), (scenario, posts, result.stderr)
        if launch:
            assert scans == 2, scans
            assert posts[0][0] == "/launch"
            eligible = [game["path"].encode() for system in ("one", "two")
                        for game in games[system]
                        if (not saved["system"] or saved["system"] == system)
                        and (not saved["favorites"] or game["favorite"])]
            assert posts[0][1] in eligible, posts
            actual = json.loads(settings.read_text())
            assert actual == {key: value for key, value in saved.items() if key != "yolo" or value}, actual
        else:
            assert (settings.read_bytes() if settings.exists() else None) == before
        assert not (state / "ui.log").exists(), "SDL initialized in direct mode"
        count += 1

    try:
        for yolo in (False, True):
            for system in ("", "one", "two"):
                for favorite in (False, True):
                    run(dict(system=system, favorites=favorite, yolo=yolo))
        saved = dict(system="one", favorites=True)
        run(saved)  # Older settings with no YOLO key.
        for system in ("gone", "ports", "tools", "emulators"):
            run(dict(system=system, favorites=False), launch=False)
        for scenario in ("empty", "hidden", "missing", "ambiguous", "stale", "moved-system",
                         "unfavorite", "scan-error", "rescan-error", "own-launcher", "own-shortcut"):
            run(saved, scenario=scenario, launch=False)
        run(saved, save_block=True, launch=False)
        for invalid in (None, "{", "null", "[]", "{}", '{"system":null,"favorites":false}',
                        '{"system":"one","favorites":null}', '{"system":"one","favorites":"yes"}',
                        '{"system":"one","favorites":true,"yolo":null}',
                        '{"system":"one","favorites":true,"yolo":"yes"}', " " * 32769):
            run(invalid, launch=False)
            assert scans == 0
        run(saved, launch=False, args=("--demo",))
        for scenario in ("launch-error", "disconnect"):
            run(saved, scenario=scenario)
        server.shutdown()
        server.server_close()
        run(saved, launch=False)  # No ES listener.

        # Test both real shell entry points in a path with spaces. A spy records
        # only the invocation contract; the cases above execute the real binary.
        (package / "state").mkdir()
        (package / "state/settings.json").write_text(json.dumps(saved))
        spy = package / "random-game-picker"
        spy.write_text('#!/bin/sh\nprintf "args:%s state:%s\\n" "$*" "$PICKER_STATE_DIR"\nexit 7\n')
        spy.chmod(0o755)
        for name, expected in (("Play Random Game.sh", "--play-random"), ("RandomGamePicker.sh", "")):
            shortcut = package / name
            shutil.copy2(root_repo / "packaging" / name, shortcut)
            result = subprocess.run(["sh", str(shortcut)], cwd="/", env=environment)
            assert result.returncode == 7
            assert (package / "state/last-run.log").read_text() == f"args:{expected} state:{package / 'state'}\n"
            assert json.loads((package / "state/settings.json").read_text()) == saved
        print(f"PASS: {count} real-binary direct-launch cases; both packaged entry points, shared state, logs and exit status.")
    finally:
        server.server_close()
