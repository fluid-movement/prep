---
title: TUI knowledge view
kind: code
parent: 20261005-152616
depends_on:
  - 20261005-152617
---

Knowledge in the TUI serves the issue workflow. Issue detail shows entries linked from the issue's context and, for finished issues, the entries its resolution changed (computed backlinks). A knowledge view lists entries by title, description and status with the same query engine (type, status, trust tier, scope) and renders the entry in the detail pane; editing via $EDITOR. An attention view lists entries flagged by the drift check, drafts awaiting review and broken links. At completion, knowledge changes are reviewed alongside other content changes in one pass.

## Open questions
