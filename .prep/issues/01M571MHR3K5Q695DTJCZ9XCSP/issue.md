---
title: 'Keep the data for the usage analysis: archive activity, record hook outcomes and failed commands'
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - activity
priority: high
---

After 0.3.0 the user uses prep for one to two weeks; then sessions, tool and token use, knowledge and hooks are analysed. Make sure that data exists: the activity stream archives instead of dropping old lines, hooks record their outcomes (prompt sent or skipped, finish asked, redirect denied and on what), and failed prep commands record their error code. Transcript retention is raised in the user's settings.

## Open questions
