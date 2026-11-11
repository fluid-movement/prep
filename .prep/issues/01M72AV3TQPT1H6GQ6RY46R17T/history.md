- 2026-11-11T08:25:01Z edited by cli/prep-v0.3.0-3-gb07f9b4: priority

- 2026-11-11T08:25:01Z cli/prep-v0.3.0-3-gb07f9b4: Review 2026-11-11: raised to high and to be implemented first among the usage follow-ups, so the hook and hand-off changes can be compared against a baseline recorded before them. Request events stay out of scope; 01M74VB2DVCENFEXXT7E1CX54A owns that gap.

- 2026-11-11T08:31:35Z cli/prep-v0.3.0-3-gb07f9b4: Open: prep-command (kind prep) events from a subagent cannot carry the agent id. The binary reads only its environment and Claude Code gives a subagent's shell no agent id; tool and hook events take it from the hook payload. Needs a decision (for example an env var the harness sets, or attributing through the Bash tool event) before criterion 2 covers prep events.

- 2026-11-11T08:32:05Z cli/prep-v0.3.0-3-gb07f9b4: Implemented, ready for check: hooks record error events, primed/sent/asked/denied carry injected chars, tool and hook events carry the Claude Code agent_id, and prep activity --all reads the archives; prep-command events from subagents do not carry the agent id (see the open note).
