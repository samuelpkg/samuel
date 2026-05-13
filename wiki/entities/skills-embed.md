---
title: skills/embed.go (binary skill bundling)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-skill-model]
tags: [v1, skill-model]
---

# skills/embed.go

The mechanism that ships every Samuel skill inside the binary.

## How it works

```go
//go:embed all:content
var content embed.FS

func FS() (fs.FS, error) {
    return fs.Sub(content, "content")
}
```

- `all:content` walks the whole `content/` tree (no glob — recursive).
- `fs.Sub` strips the prefix so consumers see `go-guide/...` not `content/go-guide/...`.
- `MustFS()` panics on error (unreachable since embed is fixed at build time).

## Where it's consumed

The package comment in `embed.go` says:

> "The orchestrator's samuel-skills component reads from this fs.FS when populating the global ~/.claude/skills/samuel/ tree."

So the flow is:
1. Build embeds `samuel_v1/internal/skills/content/**` into the binary.
2. At install/sync time, [[entities/registry]] + orchestrator copies the embedded tree to `~/.claude/skills/samuel/`.

## Historical context

Package comment (`embed.go:1-9`):

> "This is the v4 replacement for v3's 'download tarball + extract template/' flow. The orchestrator's samuel-skills component reads from this fs.FS when populating the global ~/.claude/skills/samuel/ tree."

So v3 of Samuel pulled a tarball at install time; v4 (the codebase the user calls "v1") embeds everything. v2 (the rebuild) will need to pick again.

## v2 implications

- `#open` — embed-everything vs lean-binary-with-fetch. Tradeoffs:
  - Embed: single binary, no network at install, predictable. But every skill update = rebuild + redistribute. Conflicts with "pluggable" goal.
  - Fetch: lean binary, skills versioned independently. Needs registry endpoint, caching, integrity verification.
  - Hybrid: small embedded "blessed" set + dynamic fetch for the rest.
- `#drop` — `content/` mirror at `internal/skills/content/`. v1 duplicates skill content there AND at `.claude/skills/` (verify in pass 8). Either way, v2 should have one source of truth.
