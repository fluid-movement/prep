---
type: decision
title: Brief leaves tracker checks with the orchestrator, including checker knowledge verification
description: Checker briefs omit knowledge-entry DoD items and explicitly assign them to the orchestrator; prep-command filtering, decision bounds and optional --out behavior stay unchanged.
date: "2026-11-11"
issue: 01M74V9A3E6Y9N3KMS9QGHCY1J
supersedes:
  - 01M75C2Z12Q8WCTGAKSS2J8W0C
---

Updates the brief behavior after the user-approved trial in finding 01M7648TW8B9QK8M953A8BC85P; retains the other choices of the superseded brief implementation decision.

- Definition of Done items naming a prep command (such as prep check or prep fmt --check) are omitted from every brief and checked by the orchestrator at hand-back. Checker briefs additionally omit items about knowledge entries and explicitly assign knowledge-entry updates and other tracker-only checks to the orchestrator. Reason: the trial's worktrees and title checkers could not verify knowledge updates while forbidden to read .prep, reported Problems despite meeting all criteria, and handback left those issues in progress. A delegated checker should not be asked to inspect inaccessible records. Guide's solo check pass still sees the full Definition of Done.
- The decisions section remains bounded at 6,000 bytes, with 200 reserved for notes. Decisions are taken in order: own issue, ancestors nearest first, then scoped by literal scope length. An own body may be cut when at least 200 bytes remain; cuts and omitted entries are marked. Later smaller entries may still fit. Real issues showed broad context paths could fill the bound.
- --out keeps an optional value by position: --out=<file> always names a file; --out <arg> names a file when another positional argument remains for the issue; otherwise it uses the default local hand-off location.
- Building a brief creates the hand-off directory even for stdout output, so the agent can write its report with any tool.
- Prep's role wording names no tracker command; decision texts and context remain project content and may mention commands. Only the report path is named by the brief as a tracker path.
