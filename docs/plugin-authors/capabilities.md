# Capabilities

Every plugin declares the host resources it touches. Samuel classifies each declared capability as **safe-default** (no prompt) or **risky** (interactive prompt unless `--yes`), and gates every host call against the granted set at runtime.

## Capability namespace

| Capability | Form | Classification |
| --- | --- | --- |
| `filesystem.read` | path glob | safe-default if scoped under `/workspace/**`; risky otherwise |
| `filesystem.write` | path glob | always risky |
| `exec` | command name (no shell) | risky |
| `network.outbound` | host pattern | risky |
| `samuel.api` | API method name | risky |
| `assistant.invoke` | none | risky |

Path globs use [doublestar](https://github.com/bmatcuk/doublestar) syntax (`**`, `?`, `[abc]`, …). Host patterns are domain globs (`*.openai.com`, `api.github.com`, `*`).

## The install prompt

When you run `samuel install <plugin>`, the install path:

1. Resolves the plugin and fetches its manifest.
2. Splits requested capabilities into safe-default and risky.
3. Prints the risky list and prompts:

```text
samuel-design-doc requests these capabilities:
  ✓ filesystem.read:/workspace/**         (safe-default, auto-granted)
  ! filesystem.write:/workspace/docs/**   (risky)
  ! network.outbound:api.notion.com       (risky)
  ! exec:notion-cli                       (risky)

Grant all risky capabilities? [y/N/details]
```

`details` prints a per-capability explanation (text the manifest carries in `[capabilities.<name>.reason]`). `y` grants the lot; `N` aborts install.

## `--yes`

`samuel install --yes` auto-grants every risky capability without prompting. Use this in CI or when scripting installs. Decide once on the policy; pair with a lockfile review so unexpected capabilities surface in diffs.

## Non-interactive: fail-closed

`samuel install --non-interactive` (set automatically when stdin isn't a TTY in some shells) **fails closed** if any risky capability needs granting. There is no implicit grant in non-interactive mode — pass `--yes` explicitly if that's the policy.

CI builds should always be explicit:

```yaml
- run: samuel install samuel-typescript --yes --non-interactive
```

## Runtime enforcement

The grant is recorded in `samuel.lock`. At runtime, every host call the plugin makes (`samuel.fs_write`, `samuel.net_outbound`, `samuel.exec`) goes through `HostState.Authorize`, which checks the granted set. A WASM plugin that tries to `fs_write` a path it didn't declare gets `SAM-CAP-DENY` back — there is no escape hatch.

OCI plugins are doubly enforced: the gRPC bridge gates capability-sensitive RPCs, and the container is launched with `--network none` unless `network.outbound` was granted, and with read-only mounts unless the plugin declared writes under those paths.

## Updating capabilities

`samuel update <plugin>` re-resolves the manifest. If the new version adds capabilities, the install prompt fires again — auto-update doesn't grow the trust boundary silently.

## See also

- [RFD 0003](../rfd/0003.md) — full capability model, design rationale, rejected alternatives.
- [Signing](signing.md) — the other half of plugin trust (who wrote this code).
