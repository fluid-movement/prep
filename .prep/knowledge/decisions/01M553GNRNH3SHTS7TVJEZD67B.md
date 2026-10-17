---
type: decision
title: Skip temp-prefix files on load
description: Names starting with `.prep-tmp-` are ignored in .prep and issue dirs.
date: "2026-10-17"
scope:
  - internal/mdstore
---

Names starting with `.prep-tmp-` are ignored in .prep and issue dirs.

One may belong to a write in flight or be left by a crash.
