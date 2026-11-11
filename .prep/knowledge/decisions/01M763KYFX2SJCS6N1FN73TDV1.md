---
type: decision
title: Derive worktree state without storing it
description: Added `Project.Worktree` as a JSON-ignored derived field.
date: "2026-11-11"
issue: 01M5FS0Y17Q8MK6VSC08FJX1DX
---

This keeps registry contents unchanged, and the existing registry rewrite comparison remains limited to identity fields.
