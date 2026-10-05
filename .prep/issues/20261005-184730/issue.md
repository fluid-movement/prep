---
title: 'prep edit: change title, kind, parent, dependencies and requirement'
kind: code
---

The CLI is prep's API: every change to the data goes through a command, and commands carry content, not files. The storage adapter decides how it is represented; markdown is one adapter. Today title, kind, parent, dependencies and the requirement of an existing issue can only be changed by editing issue.md.

prep edit <id> changes any of them in one write, with --json and validation before writing like every write:

- --title <text> and --kind <kind> (a valid kind).
- --parent <id> sets the parent, --parent '' removes it; the parent must exist and the tree stays acyclic.
- --depends-on <id> is repeatable and replaces the whole list; --depends-on '' clears it. Targets must exist, an issue cannot depend on itself, and no cycle may form. Depending on a done issue is allowed; depending on a dropped issue is rejected.
- --body <text> or --body-file <path|-> replaces the requirement, including its open questions, like prep new.

Flags not given leave their field unchanged; at least one flag is required. A changed requirement or kind makes a defined issue stale through the existing baseline comparison; the command reports it.

## Open questions
