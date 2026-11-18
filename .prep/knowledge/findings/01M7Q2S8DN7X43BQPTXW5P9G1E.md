---
type: finding
title: 'Cloud spike: one issue from push to merge'
description: A cloud session implemented a ready issue in two minutes; branch claude/<slug>, records merged clean, claude --cloud needs a terminal.
date: "2026-11-18"
issue: 01M7Q2AQV2EXF21424KDFS7CNY
---

# Cloud spike: one issue from push to merge

Run 2026-11-18 with prep v0.3.0 in the cloud environment from the finding "Cloud environment setup script gives a 1:1 prep install". Issue implemented: 01M7Q2AR2GWXH9ZR32MSNM9VSG (prep prints the cloud environment setup script). Session: session_01PUVjAWi89f9NRvAKjbEzmp. Result: a correct, tested implementation that merged into main as a fast-forward.

## What happened

- Start: `claude --cloud "Run prep guide <id>, implement it following the guide, log 'Implemented, ready for check' and stop. Do not complete the issue. Commit your work and push the branch."` after pushing main (8167e51). The clone had the issue as ready.
- The session claimed the issue at 10:15:47Z and logged "Implemented, ready for check" at 10:17:52Z: about two minutes of work. It pushed one commit to the branch `claude/run-prep-guide-impl-ksdhwt`; the name is generated from the prompt, so the user finds it with `git ls-remote --heads origin 'claude/*'`. It did not complete the issue, as asked.
- The commit touched six files: the Go code and tests, the README line, the command table entry, and two files of the issue's own records (`claim.md`, `history.md`).
- Merge: main had not moved, so `git merge --ff-only` worked; a trial merge in a scratch worktree was also clean. The `.prep` records from the cloud session merged without conflict because each issue's records are separate files and the session did not touch indexes.
- Checks on the merged result: `go test ./...`, `just lint`, `prep check` all clean; the printed script matches the one verified in the earlier finding.

## Answers

- **Branch**: `claude/<slug-of-prompt>-<suffix>`, one per session, pushed to origin.
- **Merging records**: clean for one session on one issue. Records live in per-issue files, so parallel sessions on different issues should not conflict; two sessions on one issue would. Not tested.
- **Prompt**: the short guide prompt worked first time because the cloud session has prep installed and the SessionStart briefing; `prep brief` was not needed. A brief would still be useful for a session without the integration.
- **Terminal**: `claude --cloud` refuses to run without an interactive terminal ("--cloud requires an interactive terminal"), so a wrapper cannot call it as a plain subprocess; it needs a pty or a visible terminal (Rex pane worked). Session ID and link are printed, along with `claude --teleport <id>`.
- **Cost**: the Claude 5-hour usage window moved from 9% to 11% across the run, which includes this session's own work; the 7-day window did not move at the reported precision. So one small issue costs at most about two points of the 5-hour window.

## Not answered yet

Parallel sessions on separate issues, whether PreToolUse hooks fire in the cloud, and the blank `request` event and `hook skipped` event seen in the earlier probe. The activity stream lives in `.prep/local` inside the VM and is not committed, so it cannot be read from the branch; it needs a question to the session itself or a run that commits the activity report.
