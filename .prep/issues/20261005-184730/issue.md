---
title: Command to add and remove issue dependencies
kind: code
---

The CLI is prep's API: every change to the data has a command, so agents and other clients never edit frontmatter by hand. Dependencies can only be set at creation today (prep new --depends-on); adding or removing one later means editing depends_on in issue.md. A write command adds and removes dependencies on an existing issue, validates before writing like every write (target exists, no self-dependency, no cycle), supports --json, and keeps depends_on in canonical form.

Proposed shape: prep depend <id> --on <id>... and prep depend <id> --remove <id>...

## Open questions

- Command name and shape: prep depend with --on/--remove, or a pair such as prep link / prep unlink?
- Is depending on a done or dropped issue allowed when adding (done is harmless; dropped is a check warning today)?

## Open questions
