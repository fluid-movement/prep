---
title: Issue priorities
kind: code
tags:
  - workflow
---

Issues differ in importance, and nothing records it: prep next lists actionable issues in ID order, and the user has to remember what matters most. Issues get a priority that orders work, so the agent picks the most important actionable issue first. The original design left priorities out of scope; this issue revisits that.

- Priority is one of four fixed levels: critical, high, medium, low. It is metadata in issue.md's frontmatter (`priority: high`), like tags. An issue without one is medium, and medium is never written, so existing issues need no migration.
- Each issue has only its own priority; children do not inherit their parent's.
- prep new and prep edit take --priority <level>. Like a tag-only edit, a priority-only edit works in any state, including resolved issues, and never makes an issue stale. An unknown level is a usage error.
- Ordering: prep next, the actionable and attention lists of prep prime, and prep list sort by priority (critical first), then by ID. In tree layouts (prep list --tree, the TUI's tree mode) siblings sort the same way under their parent.
- Filtering: prep list --priority takes one or more levels (--priority high,critical); saved views in config.yaml and the TUI filter bar accept it.
- Showing: prep list, prep show and the JSON output carry the priority; the TUI list rows and detail and the Claude Code side panel show it. Medium is shown quietly or not at all, so that only deviations draw the eye.
- The TUI's edit actions can change the priority.

## Open questions
