---
name: create-skill
description: |
  Scaffold a new Agent Skill for Samuel. Generates the directory layout,
  SKILL.md frontmatter, and the manifest entries required for the
  skill to load via Samuel's plugin loader. Use when the user asks to
  "create a skill", "add a new skill", or "scaffold an Agent Skill".
license: MIT
metadata:
  author: samuel
  version: "0.1.0"
  category: workflow
  kind: builtin
---

# Create Skill

An Agent Skill is a self-contained capability bundle: a directory with
a `SKILL.md` (this format), optional supporting prompts, and a
`samuel-plugin.toml` manifest used by the loader.

This built-in walks the user through the four decisions that make a
skill load cleanly in Samuel v2:

1. **Name** — kebab-case, unique under the registry
2. **Description** — one-paragraph "what + when" used by the loader
   to surface the skill in matching prompts
3. **Capabilities** — declared permissions (`fs:read`, `net:none`, …)
4. **Manifest** — `samuel-plugin.toml` with `kind = "skill"`

## Output layout

```
<name>/
  SKILL.md
  samuel-plugin.toml
  prompts/                  (optional)
    *.md
```

Full bodies (the interactive scaffolder, the manifest writer, the
loader integration) land with PRD 0003 / 0005. This placeholder ships
so the built-in tree is non-empty after a fresh `samuel init`.
