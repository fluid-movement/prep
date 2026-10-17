---
type: decision
title: prep stages but never commits
description: Writes stage their paths except git-ignored ones (user's own config); the user or agent commits .prep with the code.
date: "2026-10-17"
scope:
  - internal/cli/write.go
  - internal/gitx/**
---

Ignored config must never fail git add (stated); not committing reason not stated.
