## D1: Pane: side overlay when wide, strip when narrow
date: 2026-10-08

Pi has no side-panel slot: its interactive layout is fixed (transcript, widgets above the editor, editor, widgets below, footer), and replacing its layout root would depend on internals. Extensions draw through widgets and overlays. On a wide terminal the pane is a right-anchored overlay that does not take keyboard focus (`nonCapturing`, shown by a `visible` width check); it covers the right part of the transcript, which the user accepted. On a narrow terminal a widget of two or three lines above the editor shows the issue, state, step, next transition and acceptance progress, and `/prep:pane` opens the full pane as a focused overlay that Esc closes. Chosen by the user over a strip-only design.

## D2: Briefing as a system prompt section, refreshed on session events
date: 2026-10-08

A `before_agent_start` handler sets `systemPromptOptions.sections.prep` to the `prep prime --hook` output. Pi records sections in the session's first system message and appends a delta only when the text changes, so an unchanged briefing costs nothing per turn and is never compacted away. It is recomputed at `session_start` (startup, resume, new, fork, reload) and after `session_compact`, the moments Claude Code's SessionStart hook fires, and stays fixed between them; the agent learns the rest from its own prep calls.

## D3: In the agent, only the pane command
date: 2026-10-08

The user wants the agent to open and close the pane and nothing more; anything else is done in prep tui. Pi gets `/prep:pane [live|project|usage]` under the same name as in Claude Code (Pi's command names may contain a colon, as its own `/skill:name` does). No `/prep:status`, `/prep:next`, `/prep:guide` or `/prep:focus`; the pane follows the agent without pinning. This deliberately departs from parity with the Claude Code plugin.

## D4: Distribution: git source pinned to the release tag until 1.0
date: 2026-10-08

`prep setup` installs `git:github.com/fluid-movement/prep@vX.Y.Z`; a root `package.json` (private, `pi-package`, a `pi` manifest) points at `plugins/pi`. This mirrors the Claude Code marketplace pin and needs no new infrastructure; Pi clones the whole repository per install, which is accepted for now. At 1.0 the package moves to npm (lighter installs, the pi.dev gallery), decided by the user.

## D5: One copy of the inference and usage rules
date: 2026-10-08

Issue inference (`infer.ts`) exists twice already (Claude Code panel, token-ledger) and usage areas once (`classify.ts`). A Pi git install clones the whole repository, so the Pi extension imports one shared, dependency-free module directly, and the Claude Code copies, which must stay inside their plugin directories, are checked identical to it by a test, as the skill copies are. The Pi extension follows bash and powershell results and edit and write paths, including nested calls (`parentToolCallId`), and only successful ones.

## D6: Tests: pure units, an SDK session with a scripted model, and a live check
date: 2026-10-08

Pure parts (inference, usage areas, view models) run under Node's test runner with type stripping, without a build step. The extension's wiring runs in an SDK session (`createAgentSession`) with pi-ai's faux provider scripting bash calls against a real prep binary in a temporary project: the briefing section, `PREP_ACTOR`, following, refresh after compaction. Drawing is checked by rendering components at fixed widths. Each issue ends with a live check in Pi.

## D7: In the agent, only /prep for the pane
date: 2026-10-08
supersedes: D3

The user wants the agent to open and close the pane and nothing more; anything else is done in prep tui. Pi gets one command, `/prep [live|project|usage]`: without an argument it toggles the pane, with a view it opens the pane on that view. No status, next, guide or focus commands; the pane follows the agent without pinning. Pi's skill commands are `/skill:<name>`, so `/prep` does not collide with the prep skill. Claude Code moves to the same name in its own issue. This deliberately departs from parity with the Claude Code plugin as it stands.
