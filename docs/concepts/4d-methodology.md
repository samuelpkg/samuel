# 4D methodology

The 4D loop is Samuel's per-task reasoning template: **Deconstruct, Diagnose, Develop, Deliver**. It's the structure baked into the default `AGENTS.md` template, and it's what the agent walks through every time the autonomous loop hands it a new task.

## The four phases

**Deconstruct.** Restate the goal in one sentence. Enumerate the smallest set of files that have to change. If the goal can't be restated cleanly, it's underspecified — split it. If the file list is "I'll figure it out as I go," the task is too large.

**Diagnose.** Read what's there. Note what's broken, what assumptions might not hold, what tests already cover the area. Most agent failures are diagnosis failures dressed up as implementation failures — the agent jumped to coding before understanding the surrounding system.

**Develop.** Make the change with tests. Respect the guardrails (function size, file size, no commented-out code). Bug fixes carry regression tests so the same bug can't return.

**Deliver.** Run the quality checks (`go test`, `golangci-lint`, whatever's in `[methodology.ralph].quality_checks`). Summarise what changed. Commit with a conventional-commits message.

## Why this loop

Most agent loops are some flavor of "read, write, repeat." 4D adds two cheap-but-load-bearing phases: Deconstruct forces the agent to commit to a scope before touching files, and Deliver forces it to verify before claiming done. Without those, agents wander (writing more than the goal needs) and lie (declaring a task done when tests are red).

The four phases are not enforced by code — they're baked into the prompt template. The agent follows them because the template tells it to. That's enough: stating the structure in the prompt is the highest-ROI intervention in the loop.

## How it composes with Ralph

4D is the *inside* of one iteration; Ralph is the *outside* loop that runs iterations. The Ralph iteration cap (`--max-iterations`, default 20) bounds how many tasks the loop attempts in one invocation. The 4D phases happen within each one.

```text
ralph loop (bounded by --max-iterations)
├── iteration 1
│   └── 4D for task T001
│       ├── Deconstruct
│       ├── Diagnose
│       ├── Develop
│       └── Deliver  → samuel run done T001
├── iteration 2
│   └── 4D for task T002
└── ...
```

## When 4D fails

If the agent reports a task done but the quality checks fail, `quality.check` returns failure and the iteration counts as failed. After `--max-consec-fails` consecutive failures (default 3), the loop aborts. That bounds blast radius when the methodology breaks down — either the task is genuinely impossible, the quality checks are wrong, or the agent is stuck. All three need a human, and the loop yields the floor instead of grinding.
