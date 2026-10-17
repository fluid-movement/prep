---
type: decision
title: "Acceptance changes are operations, not a rewrite"
description: "Acceptance changes are operations, not a rewrite"
date: "2026-10-05"
issue: 01M46P8YDRDX76J95T3VHR41A9
---

The change carries operations (check 2, add text, opt out of an item) instead of a full new criteria list, so the markdown adapter can apply them line by line and keep any prose a human wrote in acceptance.md. Alternative: re-render the file from parsed criteria (simpler, but silently drops text).
