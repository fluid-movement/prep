---
title: Catch CI failures before they reach CI
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - release
  - adoption
---

v0.2.0 shipped while CI failed: a gofmt slip in `internal/tui/app.go` passed every local step (the tests, `prep check`, `prep fmt --check`) and was caught only by CI after the tag was pushed, and the release workflow published anyway. `just ci` already runs exactly what CI runs (`lint test check`), but nothing makes anyone run it.

Close the gaps between a local commit and CI:

- `just release` runs `just ci` instead of `go test` alone, so a release never tags code CI would reject.
- The release workflow runs the CI checks before GoReleaser (or calls the CI workflow through `workflow_call` and needs it), so a tag on a failing commit publishes nothing.
- A git pre-push hook (committed under `.githooks`, enabled by `just install` with `git config core.hooksPath`) runs `just ci`, so a push that CI would reject fails locally first. Pre-push, not pre-commit: commits stay fast, and prep stages records that are committed with the code.
- The project's Definition of Done names `just ci passes` instead of listing go test, prep check and prep fmt --check separately, so agents completing an issue run gofmt and go vet too.
- Optionally, a Claude Code hook that runs gofmt on Go files after an edit, so formatting never drifts in the first place.

## Open questions
- Pre-push hook opt-in (`just install` sets it) or opt-out?
- Should the editor-side gofmt hook live in this repository's `.claude/settings.json`, or stay a personal choice?
