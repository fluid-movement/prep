---
type: decision
title: Write permissions belong to the harness
description: prep has no approval mechanism; reads and writes are separate commands so harness rules can tell them apart.
date: "2026-10-17"
scope:
  - internal/cli
---

Stated: harness rules must distinguish reads from writes.
