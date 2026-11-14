# Original 110 base source providers (historical upstream snapshot)

Provenance: Xanerix/guoapp commit 62fb9d4e778e2f44c6ac64da7368921630b2dc5f (2026-11-11).
This directory contains the upstream public 15 native Go provider source implementation as it existed before the provider sources were moved into a separate repository. The subsequent upstream 1.0.10+110 provider MissAV (2026-11-13) is NOT included in this snapshot and must be recovered separately for exact 110 source parity.
The 13 optional custom providers live in native/core/custom_registry_sources.go, are registered into the same Go core, and have their own source IDs.
