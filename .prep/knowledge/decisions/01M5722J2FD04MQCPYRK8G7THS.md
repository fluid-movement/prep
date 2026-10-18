---
type: decision
title: Full activity streams are archived, not pruned
description: Past 1 MiB the activity stream moves to a dated archive and a new stream starts; readers see only the current stream.
date: "2026-10-18"
issue: 01M571MHR3K5Q695DTJCZ9XCSP
scope:
  - internal/activity/**
supersedes:
  - 01M553H1RJD0BF6ZHX6W4EYS08
---

Pruning dropped history after about 2,000 lines, two to three days at this rate, and the usage analysis needs weeks. Archives are never read or pruned by prep, so reads cost the same as before. The rename is atomic: concurrent writers at worst archive once.
