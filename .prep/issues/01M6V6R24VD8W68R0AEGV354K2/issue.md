---
title: Stop hook asks about decisions once three or more code files have changed since it last asked
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
---

The finish hook asks its decision question once at least three files outside `.prep` have changed since it last asked (or since the session started). A turn that leaves the count at one or two stays silent, records that it skipped, and its files carry over: the next turn is judged on everything changed since the last ask.

Why: every ask costs one full request that re-reads the session's context (about 2-3% of session cost in daypack and fpa-events), and most asks return nothing. Across 81 asks in three projects the question led to a recorded decision or finding in 3 of 26 turns that changed one or two files (12%), against 10 of 21 for three to eight files and 16 of 34 for more (finding 01M72A9TBJ5JQXT2VGT15N2HW1).

Why the count carries over (decided 2026-11-11 in review, replacing "each turn is judged on its own"): a conversational session of one-file turns would otherwise never be asked, and those are the turns in which the user states decisions. Carrying over delays the question instead of dropping it, so the decisions lost to the threshold are fewer than the one in ten estimated for a per-turn rule; the saving in asks is somewhat smaller than a third. The skipped turns are recorded with their running count, so the activity report can show whether the threshold should move.

The session rule "record a decision when it is made" stays; it produced 25 of daypack's 30 decisions without the Stop question.

Nothing else was found to separate useful asks from empty ones: whether scoped decisions exist, whether a person or a task notification started the turn, and whether the turn already recorded a decision all made no difference.

## Open questions
