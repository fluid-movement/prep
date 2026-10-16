---
title: Projects in temporary directories pollute the project registry
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - cli
  - projects
priority: low
---

Any prep command run in a directory records that directory in the machine-wide registry (`$XDG_STATE_HOME/prep/projects.json`, [Project registry](/components/projects.md)); a project created under the system temp directory shows up in `prep projects` and in the briefing's "Other prep projects here" list of every other project until the directory is deleted. Running `prep init` in a scratch directory three times left `tmp.qtkhi1kpjp`, `tmp.54rzp9az8q`, `tmp.qutneq1uty` in the real registry (removed by hand with `prep projects remove`); it happened again later in the same day. Agents see these names in every session's briefing and may try to use them.

## Open questions
