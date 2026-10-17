- 2026-10-17T07:39:13Z edited by claude-code/opus-5.5: requirement

- 2026-10-17T07:44:27Z claude-code/opus-5.5: Removed desktop/, its CI job, justfile recipes, the projects.go comment and desktop mentions in knowledge. Also removed prep watch: its last client was the desktop app (plugins never used it; the TUI uses internal/watch directly), per the TUI decision that watch stays only if something outside the TUI uses it. The desktop knowledge entry needs a git rm by the user: no prep command deletes entries and the record hook blocks it.

- 2026-10-17T07:53:11Z claude-code/opus-5.5: Removed check I020 and its contract fixture; I025 skips resolved issues; test TestRemovedEntryIsHistoryForResolvedIssues fails on the old code. go test, prep check (0 errors) and fmt pass locally. Committing and pushing to verify criterion 3 on GitHub CI.
