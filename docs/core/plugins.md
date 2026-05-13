# Plugins

Plugins are the only extension mechanism in Samuel v2. There are no first-class language, framework, or workflow concepts in the framework itself — everything is a plugin.

## Three tiers

| Tier | What it is | When to use |
| --- | --- | --- |
| **skill** | Text + scripts (a `SKILL.md` plus optional `scripts/`, `references/`, `assets/`) | Static prompts, conventions, idiomatic snippets. The majority of plugins. |
| **wasm** | A WASM module compiled with TinyGo, sandboxed by [wazero](https://wazero.io). Pure Go runtime — no host deps. | Translators, validators, formatters, anything CPU-bound that needs to be cross-platform and trustworthy. |
| **oci** | An OCI image run by Podman / Docker, talking to Samuel over a Unix-socket gRPC bridge. | Heavy tools — language servers, headless browsers, GPU workloads, anything with native deps. |

Decision tree: prefer skill if static text suffices. Step up to WASM if the plugin needs compute (parse, walk, render) but can run pure Go. Step up to OCI only when you need native libraries, a specific language runtime, or hardware access.

A fourth pseudo-tier is **meta** — a plugin with `kind = "meta"` and a `[requires]` block that installs other plugins. The [samuel-starter](https://github.com/samuelpkg/samuel-starter) pack is the canonical meta plugin.

## Lifecycle

Every plugin implements four methods (the v2 `Plugin` interface, ported from v1's `Component`):

| Method | Called by | Contract |
| --- | --- | --- |
| `Manifest()` | always first | returns parsed `samuel-plugin.toml` |
| `Install(ctx, env)` | `samuel install`, `samuel init` | writes files, returns mutation log |
| `Check(ctx, env)` | `samuel doctor` | returns healthy/unhealthy + reason |
| `Uninstall(ctx, env)` | `samuel uninstall` | reverses the mutation log |

The orchestrator replays the mutation log on uninstall — plugins don't need to re-detect what they installed. See [RFD 0005](../rfd/0005.md).

## Capability model

Every plugin declares the capabilities it needs in `[capabilities]`. The install path classifies them as **safe-default** (no prompt) or **risky** (interactive prompt unless `--yes`).

| Capability | Risky? |
| --- | --- |
| `filesystem.read:/workspace/**` | safe-default |
| `filesystem.read:/workspace/<specific path>` | safe-default |
| `filesystem.read:<anywhere outside /workspace>` | risky |
| `filesystem.write:**` | always risky |
| `exec` | risky |
| `network.outbound:<host>` | risky (host pattern) |
| `samuel.api` | risky |
| `assistant.invoke` | risky |

Non-interactive shells (`--non-interactive`) fail closed on risky capabilities. See [Capabilities](../plugin-authors/capabilities.md).

## Where plugins live

| Path | Scope |
| --- | --- |
| `<project>/.samuel/plugins/<name>/` | project-local (default) |
| `~/.samuel/plugins/<name>/` | user-global (opt-in) |
| `~/.samuel/builtins/` | embedded built-ins (`ralph`, `create-skill`, `sync`, `generate-agents-md`) |
| `~/.samuel/cache/wasm-compiled/` | wazero compile cache |
| `~/.samuel/cache/registries/<host>/` | registry index cache, 24h TTL |

## `samuel.lock`

The lockfile records what's installed, at what version, with what signature digest, and the mutation log produced by `Install`. It is machine-managed — never edit it by hand. `samuel install`, `samuel update`, and `samuel uninstall` mutate it under the project file lock. Commit it like you commit `go.sum` or `package-lock.json`.

## Built-in plugins

Four plugins ship inside the binary and are installed by `samuel init`:

- `ralph` — the methodology runtime (the autonomous loop)
- `sync` — the per-folder AGENTS.md walker
- `create-skill` — interactive scaffolder for new skill plugins
- `generate-agents-md` — the template renderer

They live at `~/.samuel/builtins/` with content-hash idempotency: `samuel doctor --fix` rewrites them when they drift from the binary's embedded copy.
