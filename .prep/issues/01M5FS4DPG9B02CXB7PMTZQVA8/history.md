- 2026-11-11T20:03:20Z edited by claude-code: requirement

- 2026-11-11T20:08:34Z claude-code: Implemented, ready for check: Added a project-specific terminal title to the TUI and balanced terminal title-stack push/pop sequences around its run. Changes are in the current working directory.

- 2026-11-11T20:08:34Z claude-code: Problems: The handoff requires updating knowledge entries describing changed behavior, but separately prohibits reading or changing `.prep` except for this report. I left the knowledge base untouched to follow that restriction.

- 2026-11-11T20:09:47Z claude-code: Check failed: The tests verify the view title and sequence writer separately, but do not exercise `Run` with a non-terminal writer or directly verify deferred pop on an error/switch. Those paths are supported by the implementation structure; terminal integration was not independently exercised.
