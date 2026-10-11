---
title: 'Link management: prep link add, rename, remove, and check'
kind: code
parent: 01M4N5ARVE6J5APZYJ7VA0N464
depends_on:
  - 01M4N5BKPZ1JAPSZC27P1272B4
tags:
  - projects
---

Links between projects are written and changed only through commands, and `prep check` keeps them honest.

- `project.md` frontmatter gains `links:`: local name → `{id, remote}` (remote optional, the clone hint). Local names follow the project-name rules and are unique within the project.
- `prep link add <name> <project>` adds a link to a registered project (by name, ID, remote or path), taking its ID and remote; the name defaults to the target's `name:`. A target without `id:` is refused with the fix to run there.
- `prep link rename <old> <new>` rewrites the link and every reference to it in issues and knowledge entries in one change. `prep link remove <name>` is refused while anything refers to it, listing the references.
- `prep links` lists the links with each target's resolution here: checked out (root, current name), unknown (clone hint).
- `prep check`: a reference to an undefined local name, or a link whose ID is not a ULID, is an error; a link whose target is checked out with a different remote than the hint is a warning that `prep fix` resolves by updating the hint (reads never write records); two links to one ID is a warning.

## Open questions
