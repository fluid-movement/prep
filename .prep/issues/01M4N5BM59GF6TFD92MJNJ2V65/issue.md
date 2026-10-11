---
title: Cross-project knowledge links
kind: code
parent: 01M4N5ARVE6J5APZYJ7VA0N464
depends_on:
  - 01M4N5BM0VSJS2Y4JE5WEH1EQ1
tags:
  - projects
---

Records and knowledge entries can link another project's knowledge entry as `api:/components/x.md`, the same way they link local entries as `/components/x.md`.

- The link resolves through the project's `links:` like a dependency; `prep check` validates it when the project is checked out (K003 for a missing entry) and leaves it alone otherwise.
- `prep guide` and `prep show` list cross-project entries an issue's context links, with the command that reads them (`prep knowledge show /components/x.md --project api`).
- The TUI and the knowledge screen show such links with the project prefix; opening one switches to that project.

## Open questions
