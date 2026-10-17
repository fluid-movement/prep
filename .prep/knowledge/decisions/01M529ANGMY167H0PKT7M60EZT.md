---
type: decision
title: "Inject only what changed; rules injected at session start"
description: "Inject only what changed; rules injected at session start"
date: "2026-10-16"
issue: 01M4ZCM7TTPASZ2EN8MQN6EPVZ
---

User agreed, 2026-10-16. The prompt hook crafts its text as now but sends it only when it differs from what it last sent in the session; session start re-fires after compaction and restores the briefing. The workflow rules (about 1.5 KB) are injected at session start; the skill is a short copy of them for setups without hooks.
