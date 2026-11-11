---
type: decision
title: Every decision names an issue, a scope or project-wide; a linked decision also applies to the files its issue changed
description: 'prep decide requires an issue, --scope or --project-wide (stored as scope **, listed in prime). Guide and brief list decisions of the issue and its ancestors. A done issue''s decisions get a scope derived from git: the files its completion changed. The Stop question proposes no scope globs.'
date: "2026-11-11"
issue: 01M5FS7KX63E8KZARSXXAS44C4
supersedes:
  - 01M72CWZZPSFV31JGEHSP22DE7
---

Decided 2026-11-11 by the agent in a review of the usage follow-ups, the user having left open questions to it. Replaces the decision of 2026-11-10 on the same subject; rules 1 and 2 are kept, the rest changes.

What the review found: the 2026-11-10 rule lets a decision name an issue instead of a scope, and guide and brief deliver a linked decision only to that issue and its descendants. A decision linked to a leaf issue therefore reaches nobody once the issue is done, and the issue link is the common route (189 of 294 decisions in prep, 29 of 40 in fpa-events, 12 of 30 in daypack, against 117, 14 and 0 with a scope). The edit-time notice works on scope alone and would name nothing in daypack. "Derive scope from the completion commit" had been rejected because the ancestor rule reaches linked decisions more simply; that holds for descendants, not for a later issue that edits the same code.

Decision:

1. `prep decide` needs one of: an issue, `--scope`, or `--project-wide`. Without any it fails; the error names the three choices and proposes the session's focused or in-progress issue when there is one, so the retry succeeds at once.
2. A decision linked to an issue applies to that issue and its descendants: guide and brief list the records of the issue and of its ancestors.
3. A decision linked to an issue that is done also applies to the files that issue changed. That derived scope is read from git and never stored: the files of the commit the resolution names as evidence, and of the commit that added the issue's resolution record, without `.prep` paths. It is cached per machine under `.prep/local`. Where git has neither commit the decision has no derived scope. An explicit scope always ranks above a derived one.
4. `--project-wide` is stored as the scope `**`, so it needs no new field and stays distinguishable from older records that have neither link nor scope. Current project-wide decisions are listed by title in the session briefing (`prep prime`), newest first and bounded; the edit-time notice leaves them out.
5. The Stop question proposes no scope globs. Directory globs are the broad scopes that already crowd one another (21 prep decisions on `internal/tui`), and with rule 3 a decision linked to the issue in work finds its files by itself. The question's text says to link the decision to the issue or give the narrowest `--scope`.
6. A good scope is the narrowest path or glob that covers the code the decision constrains. Prep does not check breadth; where many decisions match one file, the edit-time notice and the brief order them by specificity.

Rejected: a record that adds scope later (a second record type; rule 3 covers most of the gap); storing the changed paths in the resolution (derivable from git, and the completion is usually written before its commit exists).

Existing decisions with neither link nor scope stay valid and are found by search only.
