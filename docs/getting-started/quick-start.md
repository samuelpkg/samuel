# Quick Start

Sixty seconds from zero to a working Samuel project.

## Initialise

```bash
cd ~/code/my-project
samuel init
```

```text
→ writing samuel.toml
→ creating .samuel/{tasks,builtins,plugins}/
→ installing built-in skills (ralph, create-skill, sync, generate-agents-md)
→ rendering AGENTS.md (118 lines, under the 150-line CI gate)
✓ Samuel v2.0.0 initialised. Run `samuel doctor` to check health.
```

`samuel init` is idempotent — re-running it in an initialised project prints the current status and exits 0. Pass `--minimal` to skip the default starter plugins, or `--yes` to auto-grant capabilities for non-interactive shells.

## Edit `AGENTS.md`

`AGENTS.md` is the canonical context file. Most of it is rendered from `samuel.toml` and `.samuel/run/`; the only block you usually edit by hand is **Project context** (the free-form lead-in). The autogen marker tells `samuel sync` what to leave alone.

```bash
$EDITOR AGENTS.md
```

If you accidentally edit a managed section, `samuel sync --force` will rewrite it.

## Install a plugin

Browse the registry:

```bash
samuel search typescript
```

```text
samuel-typescript          v1.0.0   skill   TypeScript guardrails + idioms
samuel-react               v1.0.0   skill   React component conventions
samuel-claude-translator   v1.0.0   wasm    emits CLAUDE.md from AGENTS.md
```

Install the TypeScript skill:

```bash
samuel install samuel-typescript
```

```text
→ resolving samuel-typescript@^1.0.0 → v1.0.0
→ fetching from github.com/samuelpkg/samuel-typescript
→ verifying signature... ok (keyless OIDC, samuelpkg/samuel-plugin-release@v1)
→ requested capabilities: filesystem.read:/workspace (safe-default, no prompt)
→ installing into .samuel/plugins/samuel-typescript/
✓ samuel-typescript v1.0.0 installed
```

The lockfile (`samuel.lock`) now records the version, signature digest, and the mutations Samuel applied. Use `samuel uninstall samuel-typescript` to reverse them.

## Health check

```bash
samuel doctor
```

```text
samuel        ✓ healthy
ralph         ✓ healthy
sync          ✓ healthy
samuel-typescript ✓ healthy

4/4 plugins healthy.
Tip: install `samuel-claude-translator` to mirror AGENTS.md → CLAUDE.md (claude found on PATH).
```

`samuel doctor --fix` reinstalls anything reporting unhealthy. `samuel doctor --json` emits the same data for CI.

## Next

- Run an autonomous task loop — [Your First Task](first-task.md).
- Add cross-tool generation — [Cross-tool generation](../reference/cross-tool.md).
- Understand the methodology — [Methodology](../core/methodology.md).
