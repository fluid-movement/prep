---
type: decision
title: Finish asks once three files have changed since the last ask
description: Stop asks its decision question from three changed files outside .prep, counted since the last ask or session start; turns below that are skipped, recorded, and their files carry over.
date: "2026-11-11"
issue: 01M6V6R24VD8W68R0AEGV354K2
scope:
  - internal/cli/hook_finish.go
supersedes:
  - 01M553GN21TGH08RAXRGHBAG53
---

The finish hook asks its decision question only once at least three files outside `.prep` have changed since it last asked, or since the session started. A turn that leaves the count at one or two is not asked. It records a `finish` `skipped` hook event with the running file count and keeps the stored turn snapshot, marked pending, so the next prompt does not overwrite it and the next finish is judged on everything changed since the last ask. An ask, a turn with no changes, and a session start (including after compaction) reset the count.

Why a threshold: every ask costs one full request that re-reads the session's context (about 2-3% of session cost in daypack and fpa-events), and most asks return nothing. Across 81 asks in three projects the question led to a recorded decision or finding in 3 of 26 turns that changed one or two files (12%), against 10 of 21 for three to eight files and 16 of 34 for more (finding 01M72A9TBJ5JQXT2VGT15N2HW1).

Why the count carries over (decided 2026-11-11 in review, replacing "each turn is judged on its own"): a conversational session of one-file turns would otherwise never be asked, and those are the turns in which the user states decisions. Carrying over delays the question instead of dropping it, at a somewhat smaller saving than the third of asks a per-turn rule removes.

Accepted consequence: files the user changes by hand between turns count toward the threshold while a snapshot is pending.

The session rule "record a decision when it is made" stays. Nothing else separated useful asks from empty ones (scoped decisions existing, who started the turn, a decision already recorded in the turn). Not chosen: a small model judging the last assistant message; it adds a model call and a way to lose decisions silently. The skipped events let the activity report show whether the threshold should move.

Unchanged from the decision this replaces: turn snapshots are compared, `stop_hook_active` prevents loops, and there is no LLM judge.
