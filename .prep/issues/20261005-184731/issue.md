---
title: Commands to edit title, kind and parent
kind: code
---

The CLI is prep's API: every change to the data has a command. Title, kind and parent of an existing issue can only be changed by editing issue.md frontmatter today, although the design allows changing kind later and lists reparenting as a transition for all kinds. Write commands change each of them with validation before writing (valid kind, parent exists, no cycle) and --json output. A kind change alters the baselined requirement, so it makes a defined issue stale like a requirement edit does.

## Open questions

- One command with flags (prep edit <id> --title/--kind/--parent) or one command per field (prep retitle, prep rekind, prep reparent)? Should dependencies from the dependency command share the same shape?
- Should editing the requirement prose itself get a command (for example from a file or stdin), or stay a file edit as the define step describes?

## Open questions
