# FAQ

## Why agnostic-by-design?

Because tools come and go. A framework that hard-codes "Claude" gets stuck when users want Cursor, gets stuck again when Cursor is replaced by the next thing, and accumulates compatibility debt forever. Treating `AGENTS.md` as canonical and pushing tool-specific output into translator plugins means each tool is a self-contained, disposable integration. See [Agnostic by design](../concepts/agnostic-by-design.md).

## What happens to my existing `CLAUDE.md`?

The framework doesn't touch it. Install [`samuel-claude-translator`](https://github.com/samuelpkg/samuel-claude-translator) and the translator will overwrite `CLAUDE.md` with content derived from `AGENTS.md` on every `samuel sync`. If you'd rather keep a hand-written `CLAUDE.md`, don't install the translator — it's opt-in.

## Do I need to know Go to use Samuel?

No. To *use* Samuel you need a working shell. To *write plugins* you need Go only if you target the WASM tier (TinyGo). Skill plugins are markdown + shell scripts; OCI plugins can be written in any language that runs in a container.

## How is this different from Cursor rules?

Cursor rules are scoped to one tool. Samuel's `AGENTS.md` is canonical across every tool that knows the convention. The framework also adds an autonomous loop (Ralph), a methodology runtime (4D + hooks), a plugin lifecycle (install / check / uninstall), a signature-verified plugin registry, and a capability model — Cursor rules are just rules.

## Can I use Samuel without an LLM?

The methodology runtime and the plugin model are independent of any specific LLM. `samuel init`, `samuel doctor`, `samuel install`, and `samuel sync` are useful as project-context infrastructure with or without an agent attached. The `samuel run` autonomous loop is the only command that requires an agent.

## How does the iteration cap actually stop runaway loops?

`--max-iterations` is enforced by the loop driver, not the agent. The driver counts iterations and returns when the count hits the cap. The agent has no API to override it. See [Ralph Wiggum methodology](../concepts/ralph-wiggum-methodology.md).

## What if I distrust a plugin?

Three lines of defence:

1. **Signature** — refuses to install if the cosign signature doesn't match a trusted identity. See [Signing](../plugin-authors/signing.md).
2. **Capabilities** — risky capabilities prompt on install; granted set is recorded in `samuel.lock`. See [Capabilities](../plugin-authors/capabilities.md).
3. **Sandbox** — WASM plugins run in wazero with no host deps; OCI plugins run in containers with read-only mounts and `--network none` by default. See [Plugins](../core/plugins.md).

If you still distrust, don't install. Or pin to a forked manifest path.

## Where do plugins live on disk?

Project-local at `<project>/.samuel/plugins/<name>/`. User-global (rare) at `~/.samuel/plugins/<name>/`. Caches at `~/.samuel/cache/`. See [Plugins](../core/plugins.md).

## How is Samuel licensed?

MIT. The framework, the plugins in `samuelpkg/`, and the reusable release workflow are all MIT. Third-party plugins set their own terms; check the registry entry.

## What's on the roadmap?

The [CHANGELOG](changelog.md) tracks what shipped. Near-term:

- **v2.1** — Full `sigstore-go` integration with online Rekor verification, generated gRPC bindings for the OCI bridge, the first OCI plugin (`samuel-claude-runner`), per-plugin page richer content in this docs site.
- **v2.x** — Windows support, more translator plugins, more methodology plugins.

The [RFDs](../rfd/index.md) describe the design surface; new RFDs land as larger changes are proposed.
