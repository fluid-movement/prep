---
title: Decisions linked to a done issue apply to the files that issue changed, derived from git
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
depends_on:
  - 01M72ACGWREA9E1TFNGTJB0A98
tags:
  - knowledge
  - hooks
---

A current decision that is linked to a done issue also applies to the files that issue changed. The edit-time notice names it when one of those files is first edited, and the brief includes it when the issue's context names one of those files, in both cases after the decisions with an explicit scope.

The files an issue changed are read from git and never stored: the files of the commit its resolution names as evidence, and of the commit that added its resolution record, without paths under `.prep`. The result is cached per machine under `.prep/local`. Where git has neither commit, or the commit is too broad to say anything (more than 30 files as a starting point), the decision has no derived scope.

Why: decision 01M74V5WFVS5PK8PGJ91TENV81. Most decisions are linked to an issue and have no scope (189 of 294 in prep, 29 of 40 in fpa-events, 12 of 30 in daypack; daypack has no scoped decision at all), and a decision linked to a leaf issue reaches nobody once that issue is done. The edit-time notice (01M72ACGWREA9E1TFNGTJB0A98) would stay silent for them.

The work starts with a measurement on this repository's history, recorded as a finding: for how many linked, unscoped, current decisions a derived scope is found, how many files it has, and how many decisions then match a typical file in `internal/cli` and `internal/tui`. If fewer than a third of those decisions gain a scope, or a typical file matches so many that the five-decision cap hides the explicit ones, stop after the finding and say so in the resolution; the rest is then not built.

## Open questions
