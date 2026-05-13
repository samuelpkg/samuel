---
title: SamuelComponent (framework skill sync)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-orchestrator]
tags: [v1, orchestrator, rescue]
---

# SamuelComponent

The component that brings Samuel's own skills onto the user's machine.

## What it does

```
1. Read embedded fs.FS (from internal/skills.FS, see [[entities/skills-embed]])
2. Sync content → ~/.claude/skills/samuel/                 (global install)
3. Create symlink <project>/.claude/skills/samuel/ → ~/.claude/skills/samuel/   (project entry)
```

## Idempotency

Content-hashed. SHA-256 over `P:<path>\n<bytes>` for every file in source AND target. If hashes match AND the symlink already points at target, no-op.

## Atomic swap

```
1. Stage in sibling tmp dir: <home>/.claude/skills/samuel.tmp-<rand>/
2. If <home>/.claude/skills/samuel/ exists → rename to .bak-<shortHash>
3. Rename tmp → target
4. On success: drop backup. On failure: restore backup.
```

## Path traversal defense

`syncFS` calls `filepath.IsLocal(p)` before any write. fs.FS paths like `../etc/passwd` are rejected with a structured `*Error`.

## Symlink conflict matrix

| State at symlink path | Action |
|---|---|
| missing | create |
| symlink → target | no-op |
| symlink → elsewhere | remove + recreate |
| real file or dir | **refuse**, return Recoverable error |

Last case: never clobber user data without explicit guidance.

## v2 implications #rescue

The content-sync + atomic-swap pattern survives. Two refactors:

- **Path**: move from `~/.claude/skills/samuel/` to `~/.samuel/builtins/` (or per the namespace decision when [[entities/samuel-v2]] firms up).
- **Symlink optional**: in v2, prefer to have the framework scan its own install dir directly. Symlinks are useful for cross-tool exposure (Claude Code reads `<project>/.claude/skills/`) but should be opt-in, not required.
- **Hash store**: persist the source hash to a small file (`~/.samuel/builtins/.hash`) to avoid re-hashing the entire target on every `samuel doctor`.

## Edge cases worth preserving

- Empty global dir treated as "not installed" (handles failed-mid-sync state).
- Empty `projectDir` skips project work entirely (used by user-level `samuel doctor`).
- Lstat (not Stat) on symlink check — won't follow into the symlink target.
- Skip-symlink option for users managing the project entry themselves.

## Related

- [[entities/skills-embed]] — the source filesystem
- [[entities/orchestrator]] — the lifecycle wrapper
- [[concepts/component-lifecycle]] — the pattern
