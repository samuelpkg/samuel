---
title: Config format (TOML primary, YAML supported)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v2, v2-decision, config]
---

# Config format

v2 uses TOML as the primary configuration format. YAML is supported as a secondary option where compatibility matters.

## The decision

User preference, filed 2026-05-12. TOML is less ambiguous than YAML (no significant-whitespace gotchas), first-class in Go/Rust ecosystems (Cargo.toml, pyproject.toml, dprint config), and easier for machine-generation.

## File naming convention

| File | Purpose | Owner |
|---|---|---|
| `samuel.toml` | Project config (replaces v1's `samuel.yaml`) | User-edited, sometimes machine-modified |
| `samuel.lock` | Lockfile — resolved plugin versions, capability grants, image digests | Machine-managed (like `Cargo.lock`, `package-lock.json`) |
| `samuel-plugin.toml` | Plugin manifest at the root of each plugin repo | Plugin author |
| `rfd-index.toml` | RFD master metadata index | Machine-managed (by `create-rfd` plugin) |
| Registry `index.toml` | Plugin registry index in the `samuel-registry` repo | Curated commits |

## Exception: SKILL.md frontmatter stays YAML

SKILL.md files preserve YAML frontmatter because the [Agent Skills standard](https://agentskills.io) specifies YAML. Overriding would break cross-tool portability with 25+ AI products. See [[concepts/agent-skills-standard]].

```yaml
---
name: my-skill
description: |
  What this skill does.
license: MIT
metadata:
  category: workflow
---
```

Within SKILL.md bodies, content is markdown (same as v1).

## Why TOML works for samuel config

```toml
# samuel.toml
version = "0.1.0"
default_methodology = "ralph"

[[plugins]]
name = "go-guide"
version = "1.4.2"
kind = "skill"

[[plugins]]
name = "ralph"
version = "2.0.0"
kind = "wasm"

[[plugins]]
name = "claude-runner"
version = "1.0.0"
kind = "oci"

[methodology.ralph]
enabled = true
agent = "claude"
max_iterations = 25
quality_checks = ["go build ./...", "go test ./...", "go vet ./..."]

[guardrails]
max_function_lines = 50
max_file_lines = 300
require_tests = true

[registries]
default = "github.com/ar4mirez/samuel-registry"
```

- Array-of-tables (`[[plugins]]`) maps cleanly to plugin lists.
- Section nesting (`[methodology.ralph]`) maps to subsystem config.
- Comments are first-class.
- Cargo developers feel at home immediately.

## YAML support

`samuel-plugin.yaml` (and `samuel.yaml`) are accepted as equivalents. Useful for users porting from v1 configs or who prefer YAML. TOML is the documented default in every example.

The framework's parser tries `.toml` first, then `.yaml` as fallback. Mixing both in one project is allowed but discouraged.

## Memory note

This decision is also saved as a feedback memory (`memory/config_format_preference.md`) so future sessions inherit the preference without re-deriving it.

## Related

- [[concepts/versioning-compatibility]] — manifest + lockfile shape
- [[concepts/agent-skills-standard]] — why SKILL.md stays YAML
- [[entities/config-go]] — v1's samuel.yaml schema (to be replaced)
