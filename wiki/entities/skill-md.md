---
title: SKILL.md (skill file format)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-skill-model]
tags: [v1, skill-model, rescue]
---

# SKILL.md

The required entry file for every skill. v1 implements the [[concepts/agent-skills-standard]].

## Format

```yaml
---
name: my-skill                    # required, lowercase, [a-z0-9-], max 64, name == dir name
description: |                    # required, max 1024 chars
  What this skill does + when to use.
license: MIT                      # optional
compatibility: claude-code,cursor # optional, max 500
allowed-tools: Read,Edit          # optional
metadata:
  author: name
  version: "1.0"
  category: language              # convention for language guides
---

# Heading
Body in markdown.
```

## Validation (v1, `skill.go`)

- `ValidateSkillName`: required, lowercase only, only `[a-z0-9-]`, no leading/trailing/consecutive hyphens, ≤64 chars.
- `ValidateSkillDescription`: required, ≤1024 chars after trim.
- `ValidateSkillCompatibility`: ≤500 chars when present.
- `ValidateSkillMetadata`: cross-checks `name` field matches the directory name.

## Parsing (`ParseSkillMD`)

- Frontmatter must start at line 1 with `---`.
- Closing `---` required.
- YAML unmarshalled via `gopkg.in/yaml.v3` into `SkillMetadata`.
- Body = everything after closing delimiter, trimmed.

## Optional siblings

- `scripts/` — executable resources
- `references/` — supplementary docs
- `assets/` — templates, data, themes

Presence detected by `dirExists()` and recorded in `SkillInfo.HasScripts/HasRefs/HasAssets`.

## Scaffolding

`CreateSkillScaffold(skillsDir, name)`:
1. Creates `<skillsDir>/<name>/`.
2. Writes templated `SKILL.md` (via `GetSkillTemplate`).
3. Creates `scripts/`, `references/`, `assets/` each with `.gitkeep`.

## v2 implications

- `#rescue` — the format itself. It's the open standard. Don't reinvent.
- `#rescue` — validation logic is well-formed and self-contained. Port directly.
- `#open` — should v2 require `metadata.category` instead of inferring from the slice it's in? (Would collapse the language/framework/workflow split.)
- `#open` — `allowed-tools` field is parsed but unused in pass-1 code. Audit later passes for actual enforcement.
