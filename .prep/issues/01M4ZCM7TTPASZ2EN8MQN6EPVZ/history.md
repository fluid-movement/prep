- 2026-10-16T10:21:23Z claude-code: Measured 2026-10-16. Text agents receive: session start 2.4 KB (always); prompt hook 0.9 KB on every prompt, added to the transcript each turn; skill 11 KB, loaded almost only when the user types /prep (ledger analysis); guide 1.2 KB (open) to 5 KB (in progress); prime 2.7 KB; help 6.2 KB. Tool hooks run in about 60 ms and are silent unless they redirect. The skill is mostly a command manual (one rule is 2.5 KB of flag syntax) and about half of it describes features the rethink removes. Ledger analysis: the briefing cut knowledge file reads from 108 to 5; agents skipped prep guide in 21 of 24 completions. Activity stream, 2 days: 176 knowledge bookkeeping writes (review 70, confirm 57, update 49) for 8 completed issues; knowledge find returned 8-9 KB per query. The record redirect hook blocked a prep log command because its text mentioned the records directory (false positive). Plugin A/B (5 cases x 4 variants x 5 runs, sonnet) started as eval run 20261016-101957-23ca.

- 2026-10-16T14:06:10Z claude-code: Plugin A/B, eval run 20261016-101957-23ca, sonnet, deterministic assertions only, valid runs only (57 of 100 hit the account session limit and were excluded): none 0/9, skill-only 3/11, hook-only 12/13, current 10/10. Per case hook-only and current pass almost everything; none fails all; skill-only passes 3 of 11. The hooks (session start plus prompt) carry the behavior; the skill adds little on top. The run does not separate the session-start briefing from the prompt hook. Small samples per cell (1-4 valid runs).

- 2026-10-16T16:22:33Z edited by claude-code: requirement

- 2026-10-16T16:24:04Z edited by claude-code: requirement

- 2026-10-16T18:28:37Z edited by claude-code: requirement

- 2026-10-17T10:38:55Z edited by claude-code: requirement
