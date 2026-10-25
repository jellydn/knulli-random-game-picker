#!/usr/bin/env python3
"""Exercise the real packager with controlled readelf output, without cross-building."""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
import zipfile


ROOT = Path(__file__).resolve().parents[1]


class PackageDeviceTest(unittest.TestCase):
    def package(self, versions, *, machine="ARM aarch64", needed="libc.so.6", optimize=""):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            for directory in ("build/aarch64/assets", "packaging", "docs", "bin"):
                (root / directory).mkdir(parents=True)
            for name in ("LICENSE", "THIRD_PARTY.md", "packaging/RandomGamePicker.sh", "docs/testing.md"):
                (root / name).write_text("package fixture\n")
            (root / "go.mod").write_text("module packaging.test\n\ngo 1.26.0\n")
            (root / "build/aarch64/random-game-picker").write_bytes(b"ELF fixture")
            (root / "build/aarch64/assets/DejaVuSans.ttf").write_bytes(b"font fixture")
            (root / "build/aarch64/build-info.txt").write_text("exact build inputs\n")
            (root / "bin/file").write_text(f"#!/bin/sh\nprintf '%s\\n' '{machine}'\n")
            abi = f"[Requesting program interpreter: /lib/ld-linux-aarch64.so.1]\nShared library: [{needed}]\n"
            abi += "\n".join(f"Name: GLIBC_{version}" for version in versions)
            (root / "bin/aarch64-linux-gnu-readelf").write_text(f"#!/bin/sh\ncat <<'ABI'\n{abi}\nABI\n")
            for tool in (root / "bin").iterdir():
                tool.chmod(0o755)
            env = dict(os.environ, PATH=f"{root / 'bin'}:{os.environ['PATH']}",
                       PYTHONOPTIMIZE=optimize, GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="")
            result = subprocess.run([str(ROOT / "scripts/package-device")], cwd=root,
                                    env=env, text=True, capture_output=True)
            output = root / "build/release"
            archive = output / "random-game-picker-0.2.1-knulli-aarch64.zip"
            if result.returncode == 0:
                artifact = json.loads((output / "artifact.json").read_text())
                self.assertEqual(artifact["version"], "0.2.1")
                self.assertEqual(artifact["minimum_glibc"], max(versions, key=lambda v: tuple(map(int, v.split('.')))))
                self.assertEqual(artifact["sha256"], hashlib.sha256(archive.read_bytes()).hexdigest())
                self.assertEqual((output / "build-info.txt").read_text(), "exact build inputs\n")
                with zipfile.ZipFile(archive) as package:
                    self.assertEqual(package.read("random-game-picker"), b"ELF fixture")
            else:
                self.assertFalse(archive.exists(), result.stdout + result.stderr)
            return result

    def test_accepts_published_floor_and_numeric_ordering(self):
        result = self.package(["2.9", "2.17", "2.34"])
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_rejects_higher_symbols_even_with_python_optimization(self):
        for version in ("2.35", "2.36", "2.40"):
            for optimize in ("", "1"):
                with self.subTest(version=version, optimize=optimize):
                    result = self.package(["2.17", "2.34", version], optimize=optimize)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertIn(f"Required GLIBC_{version}", result.stderr)

    def test_rejects_missing_symbol_versions(self):
        result = self.package([])
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("No required GLIBC symbol versions", result.stderr)

    def test_rejects_wrong_architecture(self):
        result = self.package(["2.34"], machine="x86-64")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Expected an AArch64 ELF", result.stderr)

    def test_rejects_unexpected_libraries(self):
        result = self.package(["2.34"], needed="libsurprise.so")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Unexpected ELF dependencies", result.stderr)


if __name__ == "__main__":
    unittest.main()
