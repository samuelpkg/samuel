---
title: core/downloader.go + core/extractor.go
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-config-sync]
tags: [v1, install, drop, rescue]
---

# downloader + extractor

Together: fetch a versioned tar.gz from the Samuel GitHub repo, cache it, then copy files from its `template/` subdir into the user's project.

This is the **v3-era flow** that `embed.go` was supposed to replace. Both paths still exist in v1.

## Downloader (`downloader.go`)

- Fetches `https://github.com/ar4mirez/samuel` archive for a tag or `main` (= "dev" version).
- Cache: `~/.config/samuel/cache/samuel-<version>/`. Dev cache cleared every call.
- Cross-device `os.Rename` fallback to recursive copy.

Security in `extractTarGz`:

- **Path traversal**: rejects entries whose resolved path is outside `dest`.
- **Symlink validation**: rejects absolute symlink targets; rejects relative targets that resolve outside `dest`.
- **Size cap**: `MaxExtractedFileSize = 100 MB` per file. Decompression-bomb defense.

Ops exposed:

- `DownloadVersion`, `DownloadFile(version, path)`, `GetLatestVersion`, `CheckForUpdates`
- `ClearCache`, `GetCacheSize`

## Extractor (`extractor.go`)

- Copies from `<cache>/template/<path>` to `<project>/<path>`. Strips `TemplatePrefix = "template/"` (defined in [[entities/registry]]).
- `force` controls overwrite. Without `force`, existing files are skipped.
- `validateContainedPath` is the path-traversal guard used by every public method.
- `BackupFile(path, backupDir)` and `RestoreBackup(backupDir)` for safe rollback.
- Convenience helpers (`WriteFile`, `ReadFile`, `RemoveFile`, `FileExists`) all path-traversal-checked.

## v2 implications

### `#drop`

- The `template/` directory + cache prefix coupling — vestigial under the embed model. Incompatible with per-plugin transport.
- Tarball-from-Samuel-repo install — replace with per-plugin fetch (Git tag, OCI artifact, WASM module).
- Hardcoded cache path (uses hardcoded `~/.config/samuel/`, not `os.UserConfigDir()`).

### `#rescue`

- **Security checks** — path traversal, symlink validation, size cap. These belong in any code that unpacks untrusted archives. Port them.
- **Cross-device rename fallback** — small but important detail.
- **Backup/restore in extractor** — useful for `samuel upgrade` rollback. Worth keeping in v2 plugin installer.

### v2 replacement direction

Per [[concepts/plugin-format]]:

- Skill plugin → `git fetch` a tag from its repo, verify cosign signature, copy into local plugin cache.
- WASM plugin → `git fetch` the wasm module, verify hash, store in cache, instantiate via wazero.
- OCI plugin → `crane pull` (or equivalent) the image, verify cosign signature, store layer cache, run via host runtime.

In all three the security primitives in `downloader.go` carry over.
