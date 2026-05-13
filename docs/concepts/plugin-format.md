# Plugin format

Samuel chose a three-tier plugin format — skill, WASM, OCI — instead of a single mechanism. The tiers are not redundant; each exists because the others can't cover its use case without bad tradeoffs.

## Why three tiers

A **skill** is text plus shell scripts. It's the cheapest possible plugin: there's no runtime, nothing to compile, nothing to sandbox beyond what the host shell already enforces. Most plugins (the language guides, the framework guides, the methodology hints) want this and only this.

But text can't run logic. A translator plugin that mirrors `AGENTS.md` → `CLAUDE.md` per folder needs to walk a tree, read files, and write derived output. That's compute. We *could* embed that compute as shell scripts in a skill, but then it inherits the user's shell environment, can do anything `rm -rf $HOME` can do, and breaks the agnostic-by-design invariant the framework worked hard to establish.

**WASM** solves that: TinyGo compiles to a portable module, [wazero](https://wazero.io) executes it in a pure-Go sandbox with no host dependencies, and the capability system gates every host call (`fs_read`, `fs_write`, `exec`, `net_outbound`, …). The plugin can do work; it cannot escape what the manifest declared.

But WASM has limits. It can't run an LSP server, drive a headless browser, hold open a long-lived database connection, or talk to a GPU. For those, only a real OS process will do. **OCI** is that escape hatch: an image runs under Podman or Docker, talks to Samuel over a Unix-socket gRPC bridge, and gets a canonical mount layout (`/workspace` rw, `/skills` ro, `/plugin/config` ro, `/.samuel/run` ro for the CLI-mutation invariant).

## Why a TOON manifest

`samuel-plugin.toml` is TOML — every plugin author's first interaction with Samuel is reading or writing this file, and TOML is the format the broader Go ecosystem already uses (Cargo, pyproject, .goreleaser, …). The schema is strict — unknown keys are an error, not a warning, so typos surface at validate time rather than runtime.

## Why capabilities

Two reasons. First: trust. A plugin you install from the registry is code by a stranger. Capabilities turn "install this" from a yes/no decision into a granular grant: this plugin needs to read `/workspace`, that's fine; this one wants `network.outbound` to `*.openai.com`, that's a prompt. Second: documentation. The `[capabilities]` block is the plugin's API contract for what it touches — you can read it in 5 seconds and know what changes installing it will make.

Safe-default (`filesystem.read:/workspace/**`) never prompts because every plugin needs it and prompting for it would train users to mash y. Risky capabilities always prompt unless `--yes` is passed; non-interactive shells fail closed. See [RFD 0003](../rfd/0003.md) for the full classification.
