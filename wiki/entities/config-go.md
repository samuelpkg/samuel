---
title: core/config.go + samuel.yaml
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-config-sync]
tags: [v1, config, drop]
---

# core/config.go + samuel.yaml

The v1 project config layer.

## v1 schema (YAML)

```yaml
version: dev
installed:
  languages:  [go]
  frameworks: []
  workflows:  [all]
  skills:     [go-guide, initialize-project, ..., auto]
registry: https://github.com/ar4mirez/samuel  # optional, default if omitted
auto:
  enabled: true
  ai_tool: claude
  max_iterations: 25
  quality_checks: [go build ./..., go test ./..., go vet ./...]
```

File: `samuel.yaml` in cwd. Fallback: `.samuel.yaml` (hidden). YAML via `gopkg.in/yaml.v3`.

## Mirrored installed lists

Four parallel slices: `Languages`, `Frameworks`, `Workflows`, `Skills`.

- `AddLanguage(go)` writes to both `Installed.Languages` and `Installed.Skills` (as `go-guide`).
- `AddFramework(react)` writes to both `Installed.Frameworks` and `Installed.Skills` (as `react`).
- `AddWorkflow(auto)` writes to both `Installed.Workflows` and `Installed.Skills` (as `auto`).
- `MigrateLanguagesToSkills/Frameworks/Workflows` backfill the mirror for older configs.

Same triple-bookkeeping pain as [[entities/registry]]. `#drop` — collapse to one list.

## Global config + cache

- `GetGlobalConfigPath()` → `~/.config/samuel/` (hardcoded path; **breaks on Windows**, should use `os.UserConfigDir()`).
- `GetCachePath()` → `~/.config/samuel/cache/`.
- `EnsureCacheDir()` creates as needed.

`GlobalConfig` struct (file currently unused by ingested code, but defined):

```go
DefaultTemplate, DefaultLanguages, DefaultFrameworks, CachePath
```

## Get/Set surface

`ValidConfigKeys` is an allowlist of 11 keys queried by `samuel config get/set`:

```
version, registry,
installed.languages, installed.frameworks, installed.workflows, installed.skills,
auto.enabled, auto.ai_tool, auto.max_iterations, auto.quality_checks
```

`SetValue` accepts comma-separated strings for list fields.

## v2 schema (proposed, TOML)

```toml
# samuel.toml
version = "0.1.0"

[[plugins]]
name = "go-guide"
version = "1.0.0"
kind = "skill"

[[plugins]]
name = "auto"
version = "2.0.0"
kind = "wasm"

[[plugins]]
name = "claude-runner"
version = "1.0.0"
kind = "oci"

[methodology.auto]            # if auto is a methodology hook, config lives here
enabled = true
agent = "claude"
max_iterations = 25
quality_checks = ["go build ./...", "go test ./...", "go vet ./..."]

[registries]
default = "ghcr.io/ar4mirez/samuel-registry"
```

Open: methodology config — under `[methodology.<name>]` (core hook config) or `[plugins.<name>.config]` (plugin-owned)?

## Related

- [[entities/registry]] — same mirroring problem
- [[concepts/extensibility-design]] — one list, kind metadata
- [[concepts/versioning-compatibility]] — manifest + lockfile shape
