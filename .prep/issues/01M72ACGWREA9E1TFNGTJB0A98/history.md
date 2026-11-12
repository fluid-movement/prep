- 2026-11-10T08:53:28Z claude-code: Review 2026-11-10: (1) Scope coverage decides how often this fires: prep 117 of 294 current decisions have a scope, fpa-events 14 of 40, daypack 0 of 30, so in daypack it would never fire. (2) Broad scopes flood the cap: 21 prep decisions are scoped to internal/tui. Order matches by scope specificity (a file before a directory before a glob over a tree), then newest first, and say how many more match.

- 2026-11-11T08:22:30Z edited by cli/prep-v0.3.0-3-gb07f9b4: requirement

- 2026-11-11T13:54:57Z claude-code: Implemented, ready for check: before-tool hook names up to five scoped decisions per edit (once per session/agent), Stop question stops listing them, with tests, activity verb, README, decision, eval case. Remaining for user: criterion 5 (real Claude Code session; Codex delivery unconfirmed, README says Claude Code only) and criterion 8 live eval run (needs OK).

- 2026-11-11T13:55:53Z claude-code: Check passed for criteria 1-4, 6, 7, 9-11 (go test ./..., prep check, prep fmt --check all clean). Open for the user: 5 (real Claude Code session) and 8 (live eval; case eval/cases/knowledge-read/scoped-decision-at-edit.yaml with fixture cart-discount-floor exists and is sound). Not completed.

- 2026-11-12T07:00:43Z claude-code: Live verification: decision context reached model once, 2nd edit silent

- 2026-11-12T07:15:18Z claude-code: Real-world verification: with before-tool hook 3/3 satisfy the decision, without 3/3 also satisfy it (model looked up knowledge itself, prompted by the prompt hook); criterion 8 not demonstrated in real setup. /tmp/liveA

- 2026-11-12T07:15:27Z claude-code: Real-world verification: with before-tool hook 3/3 satisfy the decision, without 3/3 also satisfy it (model looked up knowledge itself, prompted by the prompt hook); criterion 8 not demonstrated in real setup. /tmp/liveA
