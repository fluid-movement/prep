- 2026-10-16T18:40:49Z claude-code: Note 2026-10-16: go.mod lists golang.org/x/sys as indirect although the lock code imports it directly; go mod tidy moves it. Settle it with this change.

- 2026-10-17T08:03:39Z edited by claude-code/opus-5.5: depends_on
