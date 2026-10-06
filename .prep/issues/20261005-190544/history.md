- 2026-10-06T05:36:45Z edited by claude-code/2.1.289: title, requirement

- 2026-10-06T05:48:05Z claude-code/2.1.289: Checked filter bar, check screen and settings in a pseudo-terminal on this repository; the stale diff is covered by TestStaleDiff since no issue here is stale.

- 2026-10-06T05:48:06Z claude-code/2.1.289: The CLI's shared loader would have raced between a reload and a check running at once; each TUI load now builds its own stores. go test -race passes for internal/tui.
