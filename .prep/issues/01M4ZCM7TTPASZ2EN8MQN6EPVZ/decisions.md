## D1: Hooks are the main guidance channel
date: 2026-10-16

User decision, 2026-10-16, backed by the plugin A/B (hooks only 12/13, skill only 3/11, none 0/9). Hooks that call prep commands are the main way agents get guidance: deterministic, and prep crafts the right text for each step. In a project without prep the hooks return nothing (verified for all five hooks). Downside: the same context is injected repeatedly; to be solved by injecting less or only what changed (open question). The skill is no longer the main carrier and stops being a command manual; commands live in prep guide and --help, which match the binary.

## D2: prep guide is the implementer's complete brief
date: 2026-10-16

User agreed, 2026-10-16. For a ready issue, prep guide <id> is the complete brief for the agent that implements it: requirement, criteria, context, linked decisions and its role (implement, log, stop; do not complete). The orchestrator's subagent prompt shrinks to running prep guide and doing the work, replacing the rule to pass the skill's rules on unchanged.

## D3: MCP parked
date: 2026-10-16

User decision, 2026-10-16. Hooks plus the CLI are good enough; no MCP server for now. Revisit only on evidence, such as agents failing on CLI flags or a harness that cannot run a shell.

## D4: Inject only what changed; rules injected at session start
date: 2026-10-16

User agreed, 2026-10-16. The prompt hook crafts its text as now but sends it only when it differs from what it last sent in the session; session start re-fires after compaction and restores the briefing. The workflow rules (about 1.5 KB) are injected at session start; the skill is a short copy of them for setups without hooks.

## D5: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16/17, are carried into release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) by 01M52ZR8564ECW8GP3ZHSXEM0A, 01M52ZRM0R2QPKK4HZE6PV62CE, 01M52ZRPZXEJDACVHTHEPX77MY, 01M52ZRSZFBMWZQJPAXVHXD7M1, 01M52ZRX0B8831K07HHVAQ6S5F. Open questions raised here were settled by the release's decision issues.
