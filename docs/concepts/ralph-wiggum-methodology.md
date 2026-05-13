# Ralph Wiggum methodology

The Ralph Wiggum methodology is named for the Simpsons character whose entire personality is "I'm helping," but who also, crucially, eventually stops. That's the load-bearing observation: bounded execution is a feature, not a limitation.

## The invariant

The autonomous loop **always** stops at `--max-iterations` (default 20). No exceptions, no escape valves, no "the agent thinks there's more to do." The cap is a foundational property of the runtime, not a configurable hint.

This matters because the failure mode of unbounded loops is catastrophic: an agent that misreads its task can spend hours making changes, an agent stuck in a debug cycle can write the same file 50 times, an agent talking to a flaky API can rack up costs without producing work. The iteration cap puts a ceiling on all of that.

## Why a hard cap

Soft caps don't work. "Stop if no progress in N iterations" requires defining progress, and the agent is the one reporting progress. "Stop if cost exceeds $X" requires accurate cost tracking, and providers change pricing. "Stop on human approval" defeats the point of an autonomous loop.

A hard iteration cap doesn't depend on the agent being honest or the provider being predictable. It's a property of the loop driver itself. The driver counts iterations; when it hits the cap, it returns. The agent has no API to override that.

## Composition with other guards

The cap is one of three guards that bound autonomous execution:

| Guard | Bounds | Default |
| --- | --- | --- |
| `--max-iterations` | total iterations per `samuel run start` | 20 |
| `--max-consec-fails` | back-to-back failed iterations before abort | 3 |
| per-hook `timeout` | seconds any single hook handler can run | 30 |

All three are required. None of them alone is sufficient. The iteration cap stops runaway loops; the consec-fails guard stops thrashing on a single bad task; the timeout stops a single hook from hanging the loop forever.

## What "iteration" means

One iteration = one trip through `before:iteration` → pick task → context generation → `before:agent.invoke` → `agent.invoke` → `after:agent.invoke` → quality checks → `after:iteration`. If the agent calls `samuel run done` (or skip / reset / enqueue) during that trip, the change is recorded; the next iteration picks up the new state.

A loop that processes 20 tasks runs 20 iterations. A loop where the agent gets stuck and calls nothing also runs 20 iterations — and then stops. From the runtime's perspective, those are the same.

## Resume semantics

`samuel run start` resumes from current `prd.toon`; it does not reset the iteration counter. The cap is per-invocation. Running it three times in a row processes up to 60 tasks total, with three opportunities for a human to check state in between. That's deliberate — the cap is not "total work the loop will ever do," it's "work the loop will do without a human looking at the result."

See [RFD 0006](../rfd/0006.md) for the design, including the CLI-mutation invariant that makes resumes safe across agent retries.
