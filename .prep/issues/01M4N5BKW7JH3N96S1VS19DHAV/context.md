Code paths:
- `internal/cli/cli.go`: `globalFlags` takes `--project` (and `PREP_PROJECT`); flags beat environment variables, and `--project` with `--root` (both flags) is a usage error. `Main` resolves it through `projects.Resolve` before the command runs and sets `a.root` to the target's root; `a.home` keeps the project the command was started in (`mdstore.Find` from `--root`/`PREP_ROOT`/cwd), when there is one and it differs.
- `internal/cli/write.go`: `reportWrite` names the target project (text: a first line `in project <name> (<root>)`; JSON: `project: {name, root, id}`) whenever `--project` chose it. Staging already follows `a.store.Root`.
- `internal/cli/activity.go`: `recordEvent` appends to the target's stream as before, and also to the home stream with `Event.Project` (the target's name) and the target prefixed `<name> › `; `activity.Event.MovesFocus` is false for events with `Project`, so a cross-project write never moves the home project's current issue.
- `internal/cli/read.go`: `prime` lists the other registered projects (`projects` in JSON, at most `--max`), with the hint how to reach them.
- `internal/cli/skill.md`: a rule on working across projects.

Knowledge: [CLI](/components/cli.md), [Project registry](/components/projects.md), [Agent activity](/features/agent-activity.md), [Agent context](/features/agent-context.md).
