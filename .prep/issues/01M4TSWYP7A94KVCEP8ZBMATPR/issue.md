---
title: CLI-first contract ergonomics and external artifact references
kind: code
parent: 01M4TST3C4ZSZDKA5WKBA3ND4H
depends_on:
  - 01M4WSRP2FYKRNT2P61T88X9TH
  - 01M4TSWXEPBGMQG271CZJBBZJP
tags:
  - cli
---

Keep CLI as the primary agent interface. Improve bounded instructions, consistent JSON/error contracts, actionable errors, and client compatibility without duplicating domain logic. Support references to external workflow artifacts without copying authoritative content. Defer MCP until an observed access or interface problem justifies a thin adapter; defer dedicated OpenSpec/Beads/Taskmaster adapters and bidirectional synchronization. Reuse 01M4J8JVW74HVE79K5QDTV930Z (guide commands), 01M4QY3ABRMJXVH3W69AVJWS50 (bounded output), and existing harness issues rather than duplicating them. Small-change shortcuts remain a future idea: measure friction first. Eval initially emphasizes reliable contract compliance, uses focused repeated comparisons to inform decisions, and introduces hard CI gates only for reliable deterministic checks.

## Open questions
