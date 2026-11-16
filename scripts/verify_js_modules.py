#!/usr/bin/env python3
"""Fail closed if staged official JS modules differ from locked upstream v1.0.1."""
from pathlib import Path
import hashlib
import json
import subprocess

ROOT = Path(__file__).resolve().parents[1] / "native" / "js_modules"
EXPECTED = {
  "heitao": {
    "manifest.json": "a5ea87f19137fdb4d770d249faae97dcda48b4c320ade9b59658b69f680897af",
    "main.js": "5f59ad4e21da7d272ee1efd76fd21e18b9f640524006325a35f5d437147b75a1"
  },
  "huangju": {
    "manifest.json": "6a0b6e6a0876be2585a3f165d16b38209794029fa799482e283d7650117c36bf",
    "main.js": "33dd8316c42f3e0fb34b189c50a6414e983a4b7b4cbe05bbbc8f24d20d6bff5b"
  },
  "jinguoduanju": {
    "manifest.json": "328426a570addcdd803962eee4edc9e4595ce9459807a2da8a50bfc04d675666",
    "main.js": "fd8076de2f7ea6d0738ef30a2f4005e0786bb248c14285fccc8c5d68f3198371"
  },
  "dsd": {
    "manifest.json": "20ae6c86a0d72e174eb8004ad32905bdefdec2a323c40a24bf65445221cc6565",
    "main.js": "265afa272ac65273ec5d01b815d7905de3566c4329b57e131f854ec8105dbbbc"
  },
  "beabox": {
    "manifest.json": "a9bad28677eefb7b454647c031694f2fb3d279dbde24a423d081349578d4928a",
    "main.js": "0daff652429bdc456f7cd873f6fbe4bdf81a33ef634618239313430f1de4ab57"
  },
  "yeguo": {
    "manifest.json": "fccfda9ddd69c70c3fd339e9f02be77a6b62737324b9f94460ff409c6d29755e",
    "main.js": "d6d98cb28188ae73a04439746381a5057e920bd4682bfd730ffdd0b39402aba1"
  },
  "huangguo-video": {
    "manifest.json": "a08630adbd41b4856b251aa1f5588bbe5d8087147c823c75a70f176d5f8ab2d4",
    "main.js": "317cc53266d1493b510e6d37cfd255a88e552224a469e0b96d41a8bfa21dc62e"
  },
  "huangguoai": {
    "manifest.json": "7439d054b4672d1dc268bacfc886650e5142531905ff0fadc363e07af67efb61",
    "main.js": "5b9e86ac42295b94b2abe1c8ee6a09cb2d86e01e45e44809f34ecb8b1cb7c5d0"
  }
}
for module_id, files in EXPECTED.items():
    for filename, digest in files.items():
        path = ROOT / module_id / filename
        actual = hashlib.sha256(path.read_bytes()).hexdigest()
        assert actual == digest, f"{path}: sha256 mismatch {actual}"
    manifest = json.loads((ROOT / module_id / "manifest.json").read_text())
    assert manifest["id"] == module_id, f"{module_id}: mismatched module id"
    assert manifest["version"] == "1.0.1" and manifest["runtime"] == "js"
    assert manifest.get("api") == 1
    assert "hosts" in manifest and manifest["hosts"], f"{module_id}: no host allowlist"
    js = ROOT / module_id / "main.js"
    subprocess.run(["node", "--check", str(js)], check=True)
    source = js.read_text()
    assert "resolveMedia(task)" in source, f"{module_id}: no playback handler"
    print(f"{module_id}: upstream sha256, manifest, syntax, resolveMedia OK")
print("All eight upstream JS modules verified. Runtime integration NOT verified.")
