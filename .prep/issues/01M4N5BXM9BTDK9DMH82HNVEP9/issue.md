---
title: 'TUI: cross-project dependencies in the detail and lists'
kind: code
parent: 01M4N5ARVE6J5APZYJ7VA0N464
depends_on:
  - 01M4N5BM0VSJS2Y4JE5WEH1EQ1
  - 01M4N5BMA0V0WEB4RM7X9XRGC6
tags:
  - projects
---

The TUI shows cross-project dependencies wherever it shows local ones: the detail's `← needs` links carry the project name (`api › 01M4… title`) and the state from that project; unknown ones (project not checked out) show the remote in the subtle tone. Opening one switches to that project on that issue, and backspace returns. The list's blocked marks count cross-project blockers.

## Open questions
