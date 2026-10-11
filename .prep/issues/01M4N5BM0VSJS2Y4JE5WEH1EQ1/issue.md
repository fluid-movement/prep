---
title: 'Cross-project dependencies: depends-on another project''s issue'
kind: code
parent: 01M4N5ARVE6J5APZYJ7VA0N464
depends_on:
  - 01M4N5BKW7JH3N96S1VS19DHAV
  - 01M4N5KBDT4SMJ8EBG33MKZKN9
tags:
  - projects
---

An issue can depend on an issue in another project: `prep new|edit --depends-on api:01M4…`, where `api` is a link (see 01M4N5KBDT4SMJ8EBG33MKZKN9) and the ID any unique suffix within that project, resolved and stored as the full ID.

- With no link of that name, `--depends-on <project>:<id>` given a registered project's name, ID or path adds the link (as `prep link add`) in the same change, so linking is never a separate chore; an unknown project fails with the registered ones listed. The target issue must exist and not be dropped when the dependency is written, as for local ones.
- Resolution goes through the registry: local name → link ID → the most recently used checkout with that ID. The other project is loaded read-only and only as far as needed (its issues' states), not on every command.
- Gates and queries: an unresolved cross-project dependency blocks like a local one (`ready`, `claim`, `--blocked`, actionable, `prep next`). A dependency whose project is not checked out is "unknown": it never blocks, and it is shown with the clone hint.
- `prep show`, `guide`, `list` and the JSON contract show cross-project dependencies with local name, project name and state. `prep check` with the target checked out reports a dependency on a missing issue (error: the target's issue store was rewritten or the link points at the wrong project) and on an issue dropped since (warning, as the dependency can no longer complete); with the target not checked out it reports nothing.
- The reverse direction ("unblocks") stays within a project; `prep show <id> --dependents` scans the registered checkouts for issues that depend on this one.
- Cycles across projects are not checked (only checked-out projects could be); documented as a limit.

## Open questions
