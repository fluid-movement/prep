- 2026-11-10T08:53:28Z claude-code: Review 2026-11-10: (1) Scope coverage decides how often this fires: prep 117 of 294 current decisions have a scope, fpa-events 14 of 40, daypack 0 of 30, so in daypack it would never fire. (2) Broad scopes flood the cap: 21 prep decisions are scoped to internal/tui. Order matches by scope specificity (a file before a directory before a glob over a tree), then newest first, and say how many more match.

- 2026-11-11T08:22:30Z edited by cli/prep-v0.3.0-3-gb07f9b4: requirement
