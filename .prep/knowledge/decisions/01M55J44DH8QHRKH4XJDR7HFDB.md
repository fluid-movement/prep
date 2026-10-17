---
type: decision
title: List columns drop by title text cut off
description: A list drops ID, then state label, then kind while its visible rows cut off more than 30 title characters (each row counted up to 10).
date: "2026-10-17"
issue: 01M55HRKGKY02FH1NZ156SRPQW
---

The issue text is what the user reads, so columns give way only when titles are actually cut; short titles keep every column even in a narrow pane. The per-row cap stops one very long title from emptying the list's columns. The choice is made once per render for the visible rows (ui.FitColumns, called from the list pane in app.go) so all rows, dimmed context rows included, share columns. Replaces the width-only rule that reserved 30 cells for the title.
