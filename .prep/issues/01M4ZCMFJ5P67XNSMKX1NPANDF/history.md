- 2026-10-16T09:45:21Z claude-code: Measured 2026-10-16: read commands take 1.7-4 s; the same reads take 0.3 s on the records alone without the code tree. Cause: every load re-fingerprints the scoped files of every evidence report, knowledge review and approval on every issue (internal/mdstore/store.go:398-430). prep makes no git calls on list, show or knowledge list. Removing the 0.3.0 contract machinery, decided in the completion and review areas, restores read speed.

- 2026-10-17T10:39:06Z edited by claude-code: requirement
