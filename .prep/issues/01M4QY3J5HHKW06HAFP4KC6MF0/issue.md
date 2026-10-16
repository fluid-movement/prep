---
title: prep eval keeps the workspaces of failing runs forever
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - testing
priority: low
---

A failing run keeps its environment (workspace, claude config, the agent's edits) so the transcript can be checked against the real files. They are created under the system temp directory and nothing removes them: after about 300 sessions of the first runs, six `prep-eval-work-*` directories hold more than 600 MB. `prep eval prune` removes old stored results but not these. Make cleanup part of the tool.

## Open questions
