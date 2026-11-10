---
type: decision
title: Every decision names an issue, a scope or project-wide; an issue's decisions also apply to its descendants
description: prep decide requires an issue, --scope or --project-wide. Guide and brief list decisions linked to the issue or an ancestor. The Stop question proposes scope globs from the changed files. No derived scope and no later scope records.
date: "2026-11-10"
issue: 01M72AV3NJCBCFD5KTB608VCGE
outcome: true
---

Decided 2026-11-10 by the agent, the user having left the choice to it.

Measured the same day: of the current decisions, 10 of 294 in prep, 8 of 40 in fpa-events and 18 of 30 in daypack have neither an issue nor a scope. Scope alone is rare (117, 14 and 0), the issue link is the common route (189, 29 and 12). The 18 in daypack are planning decisions recorded with a bare `prep decide --title`, before any file existed.

Decision:

1. `prep decide` needs one of: an issue, `--scope`, or `--project-wide`. Without any it fails and says so. The choice is made when the decision is recorded, which is the only moment the record can take it, since records are immutable. Planning decisions belong to the issue or parent they were made for.
2. A decision linked to an issue applies to that issue and its descendants: guide and brief list the records of the issue and of its ancestors. A planning decision recorded on a parent then reaches every child's implementer without a scope.
3. The Stop question, when it asks, proposes `--scope` values from the directories of the files the turn changed.
4. A good scope is the narrowest path or glob that covers the code the decision constrains. Prep does not check breadth; where many decisions match one file, the edit-time notice and the brief order them by specificity.

Rejected: deriving scope from the commit of the linked issue (needs git history at read time and covers only linked decisions, which rule 2 already reaches more simply); a record that adds scope later (a second record type for a gap that rule 1 closes at the source); accepting the gap (the brief and the edit-time notice depend on decisions being findable).

Existing unlinked, unscoped decisions stay valid and are treated as project-wide, as the README already defines an absent scope.
