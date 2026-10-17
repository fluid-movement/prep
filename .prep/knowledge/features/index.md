# features

## Entries

* [Agent activity](/features/agent-activity.md) - The local activity stream under .prep/local: what prep and harnesses record, the event format, how an agent's current issue is derived, prep focus and prep activity.
* [Agent context](/features/agent-context.md) - How agents get context without loading everything: prime at session start, guide for the chosen issue, then the files it points to; knowledge retrieval order.
* [CLI contract guidance and artifact references](/features/cli-contracts.md) - Additive lifecycle response fields, actionable reconciliation and external workflow references that never count as verification.
* [Completion requirements](/features/completion-policy.md) - Completion transition requirements and project Definition of Done guidance
* [Cross-project work](/features/cross-project.md) - How a command reaches another prep project on the machine (--project), how that shows in output and activity, and how agents learn about other projects.
* [Knowledge in the issue lifecycle](/features/knowledge-lifecycle.md) - Documentation decisions and knowledge updates during issue completion
* [Lean knowledge policy](/features/knowledge-policy.md) - Decisions, constraints and navigation first; incremental knowledge maintenance.
* [Issue lifecycle](/features/lifecycle.md) - States derived from records, transitions with their gates per kind, stale drift handling, parents, dependencies and Definition of Done.
* [Local record write safety](/features/local-write-safety.md) - Checkout-local lock held for a whole write command, atomic per-file writes and the order that writes the state-defining file last.
