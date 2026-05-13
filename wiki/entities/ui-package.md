---
title: internal/ui package (output, json, prompts, spinner)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-github-ui]
tags: [v1, ui, refactor]
---

# internal/ui

CLI presentation layer. 485 lines across four files.

## `output.go` — colored text

- **Six color tags**: success (green), error (red), warn (yellow), info (cyan), bold, dim/faint.
- **Six symbols**: `✓ ✗ ⚠ → ○ ●` for success / error / warn / info / pending / active.
- **API surface**: `Success / Error / Warn / Info / Print / Bold / Dim / Header / Section / ListItem / TableRow` + indented `*Item` variants.
- `DisableColors()` toggles `color.NoColor` globally — for CI / pipes.
- Errors → **stderr**; everything else → stdout. Consistent.
- Library: `fatih/color`.

## `json.go` — versioned envelope

```json
{
  "schemaVersion": 3,
  "command": "ls",
  "success": true,
  "data": { ... },
  "error": "..."
}
```

`JSONSchemaVersion = 3` constant carries an inline comment documenting v2→v3 changes. **This is the cleanest pattern in the UI package.** v2 should keep it.

- `PrintJSON(cmd, data)` → stdout, 2-space indent.
- `PrintJSONError(cmd, err)` → stderr.

The `command` field reflects the **invoked** command path, not the handler's hardcoded label — important for legacy-alias tracking.

## `prompts.go` — interactive prompts

- `Select(label, options)` — single choice. ▸ cyan active marker. Size 10.
- `MultiSelect(label, options, defaults)` — custom impl since promptui has no native multi-select. Toggleable `[✓]/[ ]` prefixes + "Done" sentinel.
- `Confirm(label, defaultYes)` — Y/N. Honors Ctrl+C / abort.
- `Input(label, default, validate)` — free-text with validator.
- `InputWithPlaceholder(label, placeholder)` — example hint above prompt.
- Library: `manifoldco/promptui`.

## `spinner.go` — async progress UI

- `Spinner` — indeterminate. 100ms tick. `sync.Once` on `Stop()` guards against double-close.
- `ProgressBar` — determinate. Green saucer theme, 40 char width.
- `Spinner.Success(msg)` / `Spinner.Error(msg)` — stop + print status in one call.
- Library: `schollz/progressbar/v3`.

## v2 implications

### `#rescue`

- The JSON envelope shape + schema versioning (`schemaVersion: 4` for v2 if shape changes).
- The six color tags + six symbols. Same UX taxonomy.
- Errors-to-stderr / output-to-stdout split.
- The `DisableColors()` pattern (auto-detect CI / pipes is even better).
- `sync.Once` on `Stop()` for the spinner — small but important.
- `Spinner.Success/Error` one-liner pattern.

### `#refactor` — library swaps

v1's library choices were reasonable in 2024 but the [Charm](https://charm.sh) ecosystem has overtaken them. v2 should adopt:

| v1 | v2 (recommended) |
|---|---|
| `fatih/color` | `charmbracelet/lipgloss` — theme tokens, terminal capability detection, layouts |
| `manifoldco/promptui` | `charmbracelet/huh` — native multi-select, form chaining, better UX |
| `schollz/progressbar/v3` | `charmbracelet/bubbles/{spinner,progress}` — cleaner API, integrates with bubbletea if we want TUI later |

Same caller-facing API (`Success/Error/Select/Confirm/...`), modern internals.

### `#open`

- **Theme support**. v1 has fixed colors. v2 should respect terminal color schemes + offer light/dark theme toggle. Lipgloss makes this easy.
- **`samuel doctor` as TUI**. Live-updating component checks (✓ / spinner / ✗) would be a real UX upgrade. Bubbletea model. Defer to post-v2.0.
- **JSON Lines vs single JSON**. For long-running commands (`samuel run` with many iterations), streaming JSONL is friendlier than one big response at the end. Add per-event JSON output mode in v2.
