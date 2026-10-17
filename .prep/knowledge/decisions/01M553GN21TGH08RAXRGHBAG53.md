---
type: decision
title: Finish asks for one reconciliation by heuristic
description: Stop compares turn snapshots and scoped knowledge, then asks once for issue/knowledge reconciliation; stop_hook_active prevents loops; no LLM judge, agent decides.
date: "2026-10-17"
scope:
  - internal/cli/hook_finish.go
---

Heuristic by design (stated); reason for no judge not stated.
