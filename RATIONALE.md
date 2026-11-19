# Rationale for Test & Golden Fixture Synchronization (v0.2.1)

## Context

In `ODD-001` (Section 4, `W8`), `velo` and `openpassport` share the canonical OpenPassport v1.0 golden fixture (`openpassport/templates/open-passport-template`).

During the `openpassport` v1.0.1 stabilization cycle:

1. **Canonical Template Inclusion (`M8`):** `openpassport/templates/open-passport-template/manifest.json` was rebuilt to include `.gitattributes` and `LICENSE` (12 files total instead of 10) and sealed with the updated hybrid `technicalAuthority` (`"Founding AI Product Engineer & Design Engineer · Technical Product Strategist"`).
2. **New Golden Self-Hash:** The canonical template self-hash updated from `35a9fea4f93b0da51284c57f9dbd7dab7140dc4ab702176b9f23ca2269519d92` to `b6d13a7447ba0abb914f54d338c92440acf5ab151615b5ef2260049c7b2fe5af`.
3. **Security & Contract Hardening (`B1`, `H2`, `H8`):**
   - Added `SafeJoin` path-traversal defense to `pkg/manifest/manifest.go` and `TestLoadAndVerify_RejectsPathTraversal`.
   - Removed destructive `strings.TrimSpace` from `ToLF` (`H2`) and bypassed `ToLF` on binary assets (`.png`, `.pdf`) so binary files hash as raw bytes (`TestLoadAndVerify_BinaryAssetRawBytes`).
   - Hardened `ComputeManifestSelfHashFromBytes` (`H8`) to scan the RFC 8259 `"files"` array and replace only the `"path": "manifest.json"` entry (`TestH8_StrictManifestSelfHash`).
4. **CLI Extensions (`--frontend custom` & `velo update`):**
   - Added tests in `cmd/velo/main_test.go` for `--frontend custom` (headless/Keystatic-ready custom frontend layer) and `velo update --check` (cryptographic `SHA256SUMS` release verification).
   - Zero existing assertions were removed or weakened.
