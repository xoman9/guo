#!/usr/bin/env python3
"""Fail CI unless all 16 original IDs and all 13 custom IDs are linked.

This is a static registry audit only, not verification of live websites.
"""
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
EXPECTED_ORIGINAL = {
    "hongguo": "provider_hongguo.go",
    "hanxiaoquan": "provider_web.go",
    "guipian": "provider_web.go",
    "sorani": "provider_sorani.go",
    "huangdou": "provider_huangdou.go",
    "huangju": "provider_huangju.go",
    "yeguo": "provider_yeguo.go",
    "dsd": "provider_dsd.go",
    "jinguoduanju": "provider_jinguoduanju.go",
    "jinguo": "provider_jinguo.go",
    "heitao": "provider_heitao.go",
    "beabox": "provider_beabox.go",
    "cloudfront": "provider_cloudfront.go",
    "huangguoai": "provider_huangguo.go",
    "huangguo-video": "provider_huangguo.go",
    "missav": "../sources_missav/provider.go",
}
EXPECTED_EXTRA = {
    "lujj31", "liuguo", "xiaoguo", "youguo", "meiguo", "yingguo",
    "luguo", "liguo", "juguo", "ningguo", "mangguo",
    "chengguodj", "jiaoguo",
}

def flavor_ids(path: Path) -> list[str]:
    content = path.read_text(encoding="utf-8")
    match = re.search(r'RegisterFlavor\(\s*"zhenguojian"\s*,\s*\[\]string\s*\{([^}]*)\}', content, re.S)
    if not match:
        raise AssertionError(f"Flavor registry missing at {path}")
    ids = re.findall(r'"([^"]+)"', match.group(1))
    if len(ids) != len(set(ids)):
        raise AssertionError(f"Duplicate flavor entries in {path}")
    return ids

extra = flavor_ids(ROOT / "native/sources_custom/flavor_zhenguojian.go")
legacy = flavor_ids(ROOT / "native/sources_legacy109/flavor_zhenguojian.go")
missav = flavor_ids(ROOT / "native/sources_missav/flavor_zhenguojian.go")
assert set(extra) == EXPECTED_EXTRA, f"Extra sources changed: {sorted(set(extra) ^ EXPECTED_EXTRA)}"
assert set(legacy) | set(missav) == set(EXPECTED_ORIGINAL), (
    f"Original IDs missing/extra: {sorted((set(legacy) | set(missav)) ^ set(EXPECTED_ORIGINAL))}"
)
assert len(extra + legacy + missav) == len(set(extra + legacy + missav)) == 29
for source, filename in EXPECTED_ORIGINAL.items():
    file = ROOT / "native/sources_legacy109" / filename
    if not file.is_file() or "core.Register(" not in file.read_text(encoding="utf-8"):
        raise AssertionError(f"{source} has no real registered Go source implementation: {file}")
for module in ("sources_custom", "sources_legacy109", "sources_missav"):
    assert (ROOT / "native" / module / "go.mod").is_file(), f"Missing module {module}"
print("PASS: all 16 original IDs + 13 extra IDs are present and registered (29 total).")
print("NOTICE: static registry and Go tests do NOT confirm website/API/player availability.")
