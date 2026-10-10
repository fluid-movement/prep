---
title: Audit the JSON contract against what the TUI shows
kind: research
parent: 01M4JVP89F3PGDHYETHEPS3YBB
tags:
  - desktop
---

The desktop app only talks to prep through `--json` and `prep watch`. Check that everything the TUI shows and does is reachable that way: issue list and tree with derived state, detail records, views, knowledge, activity and focus, transitions and edits. Check whether `prep watch` output is enough for incremental refresh or forces full reloads. Findings list the gaps, each as a follow-up issue.

## Open questions
