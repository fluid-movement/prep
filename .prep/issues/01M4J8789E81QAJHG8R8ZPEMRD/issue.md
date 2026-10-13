---
title: Deferred Claude Code hook and behavior verification
kind: code
parent: 01M498VQD4Q85CZYK965WW3H9Q
depends_on:
  - 01M4J8780KZRKWHJR4NN8AD85M
  - 01M4SX94VFDXCRYWNKCJG1H01J
tags:
  - integration
  - adoption
priority: high
---

Verify the shared prep lifecycle hooks in real Claude Code sessions when usage is available again. The shared mechanics and Codex behavior are the current 0.3.0 focus; this issue is deferred to 0.4.0 and no Claude model calls are currently authorized. Validate plugin loading, session/prompt/Stop guidance, direct-record redirects, activity attribution and fail-open behavior. Use the agreed tracking policy: implementation is tracked regardless of size unless the user opts out; questions, exploration and planning do not automatically create issues.

Retain the original named-model validation requirements from the first behavior findings below. These are deferred, not satisfied by Codex runs:

01M4QY30FKG3RNFSREWCZCPF9W — Agents skip the tracker for tasks that do not mention prep
- natural-fix-matching-issue passes on haiku in at least 4 of 5 runs and still passes on sonnet
- natural-fix-no-issue still passes on haiku (the agent creates no issue and claims nothing when none matches)

01M4QY31SZ0W4C77J6PXYYC05T — Agents file a new issue before looking for one that already covers it
- real-bug-report-to-issue passes on haiku in at least 3 of 5 runs and on sonnet in at least 4 of 5

01M4QY32YXYAJ4TZN0VZN1ZC7H — Agents change the first matching issue when a reference is ambiguous
- adversarial-ambiguous-reference passes on haiku in at least 3 of 5 runs and on sonnet in at least 4 of 5 with the chosen change
- The cases that change issues by words (batch-bookkeeping, issues-file-child) still pass at the same rates

01M4QY36MC01A53N0MM0S55FCK — Agents leave knowledge drifting after changing the code it describes
- natural-fix-matching-issue passes the check and knowledge assertions on haiku in at least 3 of 5 runs

01M4QY3942ZC72MPNBF2Q3ETXT — Answers about effort and priority are guessed from titles, not read from the tracker
- real-plan-next passes on sonnet in at least 4 of 5 runs and on haiku in at least 4 of 5

01M4QY3DTDMRSSE90EX9P6S4SK — Agents over-reach after being asked to update a requirement
- The case ran with at least 5 runs on haiku and on sonnet and the number of runs that went beyond the request is recorded in the findings

Also rerun the retained parent cases on haiku and sonnet after the fixes and record per-case pass rates and token costs. Update natural-fix-no-issue to require a new issue for unmatched implementation under the accepted policy, with an explicit opt-out case requiring no issue. Do not run the superseded no-issue expectation unchanged. Compare the original requirement-edit scope case as well as the Stop diagnostic. Other harness integrations depend on the shared hooks, not on this later Claude verification.

## Open questions
