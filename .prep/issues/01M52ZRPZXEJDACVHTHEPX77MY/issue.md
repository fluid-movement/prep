---
title: Prompt hook sends only what changed
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZRM0R2QPKK4HZE6PV62CE
tags:
  - integration
---

The prompt hook crafts its text as now but sends it only when it differs from what it last sent in the session (state under .prep/local); session start restores the briefing after compaction. Needs an eval case showing tracking behavior holds. Decided in 01M4ZCM7TTPASZ2EN8MQN6EPVZ.

## Open questions
