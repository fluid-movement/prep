Knowledge: [Markdown store](/components/markdown-store.md), [Storage format](/conventions/storage-format.md), [Domain package](/components/domain.md) (diagnostic codes).

- Record files with `##` sections: `issue.md` (prose and `## Open questions`), `acceptance.md` (criteria and `## Definition of Done`), and the free-form `context.md` and `findings.md`. `decisions.md` headings are entries (`## D<n>: …`, checked by I019), and `history.md` has none, so neither is in scope.
- `splitOpenQuestions` (`internal/mdstore/records.go`) takes only the first `## Open questions` heading. A second copy stays in the prose, so questions listed under it never reach the `G_OPEN_QUESTIONS` gate. `parseAcceptance` re-enters the DoD section at every DoD heading, so a duplicate there loses nothing.
- Per-file diagnostics come from `Store.loadIssue` (`internal/mdstore/store.go`); codes live in `internal/domain/diag.go` (next free: I028). `Store.Fix` (`write.go`) applies safe repairs file by file through `s.write`.
- Contract corpus: `testdata/contract/{valid,invalid}/<case>/` with an `expect` file listing codes. Invalid trees must produce at least one error.
