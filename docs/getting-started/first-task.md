# Your First Task

Samuel's autonomous loop reads a PRD (Product Requirements Doc), drives your AI assistant through it one task at a time, and stops when the queue is empty or the iteration cap fires.

## 1. Write a PRD

A PRD is markdown with a list of tasks. Save it at `.samuel/tasks/0001-prd-feature.md`:

```markdown
# Add /healthz endpoint

## Goal
Expose a JSON liveness probe at `/healthz`.

## Tasks
- [ ] T001: Add handler in `internal/http/health.go`
- [ ] T002: Wire route in `internal/http/router.go`
- [ ] T003: Add table-driven test in `health_test.go`
- [ ] T004: Document the endpoint in README
```

## 2. Convert to TOON

```bash
samuel run convert .samuel/tasks/0001-prd-feature.md
```

This produces `.samuel/run/prd.toon` — a token-efficient TOON serialisation of the same task list. Samuel uses TOON for everything in `.samuel/run/` that the agent reads on every iteration; markdown stays for prose-heavy journals (`progress.md`).

## 3. Start the loop

```bash
samuel run start --ai-tool claude --max-iterations 20
```

```text
→ acquiring .samuel/.lock
→ loading .samuel/run/prd.toon (4 tasks, 0 done)
→ iteration 1/20: T001 — Add handler in internal/http/health.go
   ▶ claude code (sandbox: none, prompt: file-arg)
   ✓ agent reports done; running quality.check hooks
   ✓ go test ./... passed
→ iteration 2/20: T002 — Wire route in internal/http/router.go
   ...
✓ queue empty after 4 iterations
```

The agent receives the PRD, the task context, a project snapshot, and the rendered prompt template. It then calls `samuel run done T001` (or `skip`, or `enqueue`) to update state — it never edits `prd.toon` directly. This is the **CLI-mutation invariant** that makes the runtime reliable across agent retries.

## 4. Inspect state

State lives under `.samuel/run/`:

| File | Format | Contents |
| --- | --- | --- |
| `prd.toon` | TOON | task list (id, title, status, deps) |
| `task-context.toon` | TOON | next-task slice the agent sees |
| `project-snapshot.toon` | TOON | repo layout + detected stacks |
| `progress.md` | Markdown | append-only iteration journal |
| `progress-context.md` | Markdown | last-N-lines summary, rotated at 500 lines |

```bash
samuel run status
samuel run tasks --status pending
```

```text
loop:    idle
prd:     .samuel/run/prd.toon
tasks:   4 total — 4 done, 0 pending, 0 in-progress
last:    iteration 4 at 2026-05-13T10:42:11Z
```

## 5. Iteration cap and abort

`--max-iterations` is the **Ralph Wiggum** safety belt — the loop stops there even if the agent thinks there's more work. `--max-consec-fails` (default 3) aborts when quality checks fail back-to-back. `--profile` prints per-hook timings. `--dry-run` runs every hook except the agent call so you can verify your prompt template.

## 6. Resume

`samuel run start` resumes from the current `prd.toon` — it does not reset the queue. To reset a task, run `samuel run reset T002`. To clear in-flight state entirely, delete `.samuel/run/` and re-convert your PRD.
