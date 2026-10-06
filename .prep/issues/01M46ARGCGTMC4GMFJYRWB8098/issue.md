---
title: TUI editing and transitions
kind: code
parent: 01M46AREE07YNNVPBA8F4MT56D
depends_on:
  - 01M46ARFD8BQNT8QBJYRB30C35
tags:
  - tui
---

The TUI becomes a client that can change the data, not only show it. Every change goes through the same domain planning and validation as the CLI's write commands, recorded with a human actor, so the TUI cannot produce anything the CLI would reject.

- An action menu on the selected issue lists every action with its key; actions that do not apply now are shown with the reason from the domain's gates instead of being hidden.
- Create an issue (top-level, or as a child of the focused parent), edit the title inline, and edit the requirement and the context in $EDITOR. The edited text is validated when the editor closes; when the write is rejected the error is shown and the edited text is kept for the next attempt.
- Reparent an issue by picking the new parent from a filterable list, or make it top-level.
- Check and uncheck acceptance criteria.
- Transitions that are the human's to make: define, mark ready, acknowledge a stale change, drop with a reason in any unresolved state, and complete manual, research and decision issues with a documentation decision. Code issues are completed by an agent with commit evidence, so the TUI does not offer it.
- Copying an issue ID to hand to the agent stays the way to start work; claiming stays with the agent.

## Open questions
