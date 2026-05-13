# Guardrails

Guardrails are project-wide rules Samuel injects into every rendered `AGENTS.md`. The agent sees them on every iteration, so they shape *all* generated code — not just code at the boundary.

## Default ruleset

`samuel init` writes these to `samuel.toml`:

```toml
[guardrails]
max_function_lines = 50
max_file_lines     = 400
require_tests      = true

# Conventional commits and "no secrets, no commented-out code,
# no untracked TODOs" rules ride as boilerplate inside the
# AGENTS.md template — they aren't toggleable.

extra = [
  # Free-form project-specific rules. Each line becomes a bullet
  # in the rendered AGENTS.md Guardrails section.
]
```

## How they reach the agent

The template at [`template/AGENTS.md.tmpl`](https://github.com/samuelpkg/samuel/blob/main/template/AGENTS.md.tmpl) walks the `[guardrails]` block and emits a Guardrails section. The marker pair tells `samuel sync` what to overwrite:

```markdown
<!-- SAMUEL_GUARDRAILS_START -->
## Guardrails

- Function ≤ **50** lines. Split with helpers when you cross.
- File ≤ **400** lines.
- New code carries tests. Bug fixes carry regression tests.
- Validate inputs at every boundary. Parameterise queries. No magic
  numbers, no commented-out code, no TODOs without a tracker.
- Conventional commits: `type(scope): description`. One logical change
  per commit.
- Use repository-pattern in `internal/store/`. No raw `database/sql`
  outside that package.
<!-- SAMUEL_GUARDRAILS_END -->
```

The last bullet came from `extra`. Anything you add to `extra` is preserved across regeneration.

## Per-folder overrides

`samuel sync` walks the project tree writing per-folder `AGENTS.md` files. Each can override guardrails by placing a `samuel.toml` (with a `[guardrails]` block) in that folder. The walker merges parent + child; the child wins on conflict.

```text
my-project/
├── samuel.toml                  # max_function_lines = 50
├── AGENTS.md
└── frontend/
    ├── samuel.toml              # max_function_lines = 30  (stricter)
    └── AGENTS.md                # rendered with 30
```

## Why these defaults

50 lines per function and 400 per file are empirically painful enough that agents notice when they're crossed and naturally factor things smaller. They are not aesthetic preferences — they are token-budget heuristics. Bigger functions waste context on every iteration.

Conventional commits are non-negotiable because Samuel's release tooling (and the `samuel-plugin-release` reusable workflow) reads commit messages to compute SemVer bumps. A non-conforming history breaks plugin releases.

The "no secrets, no commented-out code, no untracked TODOs" rules ship as template boilerplate rather than toggles — turning them off is a code smell, not a feature.

## Validating the rules ran

```bash
samuel sync --dry-run
```

prints the rendered `AGENTS.md` to stdout without writing. CI can diff that against the committed file to catch drift:

```yaml
- run: samuel sync --dry-run > /tmp/agents.expected
- run: diff -u AGENTS.md /tmp/agents.expected
```

The Samuel repo itself runs this check on every PR.
