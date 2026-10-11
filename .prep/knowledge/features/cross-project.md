---
type: feature
title: Cross-project work
description: How a command reaches another prep project on the machine (--project), how that shows in output and activity, and how agents learn about other projects.
generated:
  by: claude-code/opus-5.5
  at: 2026-10-11T10:15:38Z
scope:
  - internal/cli/cli.go
---

# Cross-project work

Applies when changing how prep works with projects other than the one a command was started in. Projects are found through the [project registry](/components/projects.md); the feature is 01M4N5ARVE6J5APZYJ7VA0N464.

- **`--project`**: a global flag (or `PREP_PROJECT`) taking a name, ID suffix, remote or path (`projects.Resolve`). `resolveProject` runs before the command: it sets `a.root` to the target's root, keeps the target in `a.target`, and records the project the command was started in (`--root`/`PREP_ROOT`/cwd) as `a.home` when it differs. Flags beat environment variables: `--root` overrides `PREP_PROJECT`, `--project` overrides `PREP_ROOT`; both flags are `E_USAGE`. Ambiguous and unknown projects fail before the command runs, with the candidates or the registered names.
- **Every command works there**, reads and writes alike; writes stage in the target's repository (staging follows `a.store.Root`) and never commit. `reportWrite` names the target: a first line `in project <name> (<root>)`, JSON `project: {name, id, root}`.
- **Activity**: the event goes to the target's stream as usual (so an agent's focus there follows its writes), and a copy to the home stream with `Event.Project` set and the target prefixed `<name> › `. Events with `Project` never move focus ([Agent activity](/features/agent-activity.md)), so a write elsewhere leaves the home project's current issue alone.
- **Discovery**: `prep prime` lists the other registered projects (JSON `projects`, at most `--max`, the project's own clones left out) with how to reach them; the skill's rule 11 teaches `prep projects` and `--project`, and never to edit `id:`.
