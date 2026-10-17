## D1: Claude Code is the reference harness; Codex kept unused; planned harness work dropped
date: 2026-10-16

User decisions, 2026-10-16. Keep install.sh, prep update and prep setup. Claude Code is the reference harness, for daily use and evals. The Codex integration stays because it exists, but is not used day to day: Codex did the 0.3.0 work only because the Claude Code quota was used up, and the user was not happy with the result. Later use: hand off ready issues to Codex as an implementer to test whether the handoff design works across harnesses. The skill has one source; harness copies are generated from it. Drop when actionable work is created: the Gemini, Cursor, OpenCode, Copilot and Pi hook issues and their parent (01M4J87R4AFB193QF0G3AECNA0), eval adapters for other harnesses (01M4QKVDGN3GH5A08ZV3ZTDSX8), Homebrew tap (01M4869P6R226F3GCQKAEC5F8J), cloud session install hook (01M48RV2FR0XQAGCKEXRJFH7AE), CI check for hand edits (01M48RV4E8J5T57FN2P09SHG2W), catching CI failures before CI (01M4G30VB0SNZT4AJ16XR0FD88) and deferred Claude Code hook verification (01M4J8789E81QAJHG8R8ZPEMRD). Park the Jev classification research (01M4J8784Q2JR9V8EWAH0C0A3F) with the future knowledge relevance work. New harnesses are added as thin adapters over the shared hooks when actually used.

## D2: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16: keep install.sh, prep update and prep setup; nothing to build in release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) beyond the removals and the prune sweep.
