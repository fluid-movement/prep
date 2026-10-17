---
type: decision
title: Write pipeline holds one lock
description: 'lock, load, Plan, CheckWrite, Apply, stage: the checkout lock covers load through the last write.'
date: "2026-10-17"
scope:
  - internal/cli/write*.go
---

Each file is written through a temporary file and a rename. Holding the checkout lock from load through the last write prevents lost updates (implied).
