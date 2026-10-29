# Zhenguojian APK 0.2.391+2414: 45 new stations native recovery (round 3)

## Strict scope
- **Do not modify or re-port** working `taoguo` (桃果), `liuguo` (榴果), or `xiaoguo` (宵果).
- Recover the remaining 45 source identities and category data from the user-supplied APK.
- New UI source labels get one trailing **`🔞` glyph**, not two characters `18+`.
- This document is **static binary extraction**, not proof that an external site, category URL, or stream is currently operational. Do not activate sources without verified adapters.

## Evidence
- APK: `zhenguojian-0.2.391+2414-arm64-v8a(1).apk`
- APK SHA256: `79dda2f076375a9cfb8fa8dca221f9b8a62ed5babc139e04c9420a51a7b2998d`
- ELF: `lib/arm64-v8a/libduanju_core.so` (Go 1.26, stripped with retained gopclntab)
- Go category-table initializer: `duanjuapp/native/core.map.init.17` in `0x2ba20c0..0x2ba4a40`.
- Category array address + length recovered from the `ADRP/ADD` copy and following Go map-insert key; two UTF-8 strings per category reconstructed using `R_AARCH64_RELATIVE` relocations.
- **24 source category arrays**, **566 category pairs** in total, including completed or out-of-scope sources.
- Exactly **436 category pairs** belong to 19 of the remaining 45 sources. The other **130** belong to out-of-scope keys: `heguo` (10), `shiguo` (19), `fanguo` (36), `taoguo` (28), `liuguo` (37).
- Remaining **26 sources may use dynamic category discovery**; zero static pairs does NOT mean a provider has no categories.

## Source ID corrections
The prior string-based extraction mistook provider aliases for the native registry IDs:

| UI source | Old string | Correct native registry ID |
| --- | --- | --- |
| 倉果 🔞 | `hsck` | `cangguo` |
| 巴果 🔞 | `javbus` | `baguo` |
| 杰果 🔞 | `jable` | `jieguo` |
| 檔果 🔞 | `javdb` | `dangguo` |
| 漢果 🔞 | `hanime` | `hanguo` |

## Native static categories by source
Source names and ID/label arrays are grounded in the APK's compiled Go map initializer. They are **not** directly runnable playback adapters.

| Source ID | Category count |
| --- | ---: |
| `luguo` | 10 |
| `juguo` | 12 |
| `ningguo` | 12 |
| `cg69` | 13 |
| `chunguo` | 14 |
| `yeshe` | 15 |
| `zaoguo` | 16 |
| `chaoguo` | 16 |
| `cg51` | 19 |
| `yanguo` | 20 |
| `caoguo` | 20 |
| `youguo` | 22 |
| `huiguo` | 23 |
| `jiaoguo` | 24 |
| `zhenguo` | 24 |
| `liaoguo` | 25 |
| `yangguo` | 29 |
| `bangguo` | 30 |
| `xueguo` | 92 |
| **Total** | **436** |

## Next tasks
1. Recover dynamic category and paging routes for the other 26 sources from the corresponding provider-specific functions in the same APK.
2. Validate the source routing, covers, detail and playback on actual site HTML/API; do not infer playability from static catalog metadata.
3. Port each *verified* provider to an isolated adapter in `guo` with no upfront catalog preload; preserve all completed providers.
4. Match the native registry IDs exactly and suffix display names `🔞` once.
5. Build and test an ARM64 coexist APK only after integration tests pass.

The full 45-station CSV, 436 category pairs with per-record evidence VA, 130 excluded pairs, JSON and disassembly are attached in the conversation artifact package (not committed to this text-only repository document).
