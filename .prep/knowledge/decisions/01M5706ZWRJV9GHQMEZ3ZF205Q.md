---
type: decision
title: Handoff prompts say who checks; the check step follows role markers
description: The implement prompt tells a handed-off agent to log ready for check and stop; 'ready for check' and 'Check failed' log lines decide the check step
date: "2026-10-18"
issue: 01M56ZKEHAQE5C4G8J0NQA0KPX
scope:
  - internal/domain/guide.go
  - internal/cli/rules.md
supersedes:
  - 01M56ZKEQ4SYQTTK29AP0V9DJ1
---

Refines the decision it supersedes (2026-10-18). With implementer and checker as passes, a handed-off agent cannot tell it is not alone, so the implement handoff prompt says it: run prep guide <id>, implement it, log ready for check and stop; the review prompt stays 'run prep guide <id> and follow its checker section'. The check step follows the latest role marker in the work log, 'ready for check' from the implementer or 'Check failed: ...' from the checker sending it back, and ignores other lines, because a reviewer's note (or an accidental probe line, which happened in the first review) must not end the check. Prep still never runs agents; delegation stays the orchestrator's judgment.
