---
title: Issue priorities
kind: code
tags:
  - workflow
---

Issues differ in importance, and nothing records it: prep next lists actionable issues in ID order, and the user has to remember what matters most. Issues get a priority that prep next, prep prime, prep list and the TUI use to order work, so the agent picks the most important actionable issue first. The original design left priorities out of scope; this issue revisits that.

## Open questions

- Levels like Jira (for example critical, high, medium, low) or a ranked order between issues?
- What is the default for issues without a priority, and do children inherit their parent's priority?
- Does priority only order lists, or can it also filter (prep list --priority high) and appear in saved views?
- Is changing a priority free at any time, including on resolved issues, like tags?
