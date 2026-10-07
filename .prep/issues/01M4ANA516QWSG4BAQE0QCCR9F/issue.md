---
title: Negate query flag values; hyphens in every value
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - cli
  - views
---

A value of a list query flag can be negated with a leading `not-` or `!`: `--state not-open` lists every issue that is not open. `not-` is what the CLI documents, since shells expand an unquoted `!`; `!` is handy in the TUI filter bar and in views. Both work everywhere queries are read: `prep list`, saved views, the TUI filter bar, for `--state`, `--kind`, `--tag`, `--priority` and `--under`.

- Within one flag, negated values exclude and plain values include: `--state not-open,not-done` is neither open nor done; `--state ready,in-progress,not-dropped` is ready or in progress. An issue matches a flag when it matches one of the plain values (or there are none) and none of the negated ones.
- `--tag not-cli` matches issues without the tag; `--under not-<id>` excludes an issue's descendants.
- Tags cannot start with `not-`, so negation is never ambiguous; negated values are validated like plain ones (`--state not-nope` is an error).
- Every option value uses hyphens: the state in_progress becomes `in-progress` everywhere it appears (query values, `--json` output, views, the Claude Code panel, docs). The underscore spelling is no longer accepted; prep has no outside users, so this repository and the fixtures are converted.
- `prep flags` shows `[not-]` for list flags, and the filter bar picker offers values after `not-` or `!` as well.

## Open questions
