---
type: feature
title: Cross-project work
description: How a command reaches another prep project on the machine (--project), how that shows in output and activity, and how agents learn about other projects.
generated:
  by: claude-code/opus-5.5
  at: 2026-10-11T10:15:38Z
scope:
  - internal/cli/cli.go
  - internal/cli/links.go
  - internal/projects/links.go
  - internal/domain/links.go
confirmed_commit: 443b7c8c83a82d0359d8999aef69beedd7226633
---

# Cross-project work

Applies when changing how prep works with projects other than the one a command was started in. Projects are found through the [project registry](/components/projects.md); the feature is 01M4N5ARVE6J5APZYJ7VA0N464.

- **`--project`**: a global flag (or `PREP_PROJECT`) taking a name, ID suffix, remote or path (`projects.Resolve`). `resolveProject` runs before the command: it sets `a.root` to the target's root, keeps the target in `a.target`, and records the project the command was started in (`--root`/`PREP_ROOT`/cwd) as `a.home` when it differs. Flags beat environment variables: `--root` overrides `PREP_PROJECT`, `--project` overrides `PREP_ROOT`; both flags are `E_USAGE`. Ambiguous and unknown projects fail before the command runs, with the candidates or the registered names.
- **Every command works there**, reads and writes alike; writes stage in the target's repository (staging follows `a.store.Root`) and never commit. `reportWrite` names the target: a first line `in project <name> (<root>)`, JSON `project: {name, id, root}`.
- **Activity**: the event goes to the target's stream as usual (so an agent's focus there follows its writes), and a copy to the home stream with `Event.Project` set and the target prefixed `<name> › `. Events with `Project` never move focus ([Agent activity](/features/agent-activity.md)), so a write elsewhere leaves the home project's current issue alone.
- **Knowledge links**: `[text](<link>:/path.md#section)` in records and entries (`domain.XKnowledgeLinks`, code skipped). `prep show` lists an issue's (`knows: …` with the read command; JSON `cross_knowledge`), `prep guide` adds them as inputs, and `prep check` reports X007 (a warning) when the checked-out project has no such entry. `--project` accepts the home project's link names before registry names (`projectNamed`), so `prep knowledge show <path> --project <link>` works as references write it.
- **TUI**: the header names the project; `P` or a click on it switches projects ([TUI navigation](/components/tui-navigation.md)); `prep tui --project <name>` opens on one. The detail shows cross-project dependencies and knowledge links and follows them into their project ([TUI detail links](/components/tui-detail.md)); the list's blocked marks count cross-project blockers.
- **Discovery**: `prep prime` lists the other registered projects (JSON `projects`, at most `--max`, the project's own clones left out) with how to reach them; the skill's rule 11 teaches `prep projects` and `--project`, and never to edit `id:`.
- **Links**: project.md `links` maps each local name to the target's permanent ID and its last known remote (the clone hint). Records refer to `<name>:<full issue ID>` (dependencies) and `[text](<name>:/path.md#section)` (knowledge links); `Tree.LinkRefs` finds them in requirements, context, findings, decisions, criteria, work logs and knowledge entries. Commands (`internal/cli/links.go`): `prep links` (each link's checkout here, else its clone hint), `prep link add <project> [--as name]` (a registered project with an ID; the name defaults to its own), `prep link rename <old> <new>` (project.md and every reference in one change), `prep link remove <name>` (refused while referenced, listing where). Checks: X001 a reference to an undefined name, X002 an invalid link (name, ID, or the project itself), X003 two links to one project, X004 a stale clone hint (`prep fix` updates it).
- **Dependencies on other projects**: `prep new|edit --depends-on <project>:<ID or suffix>` (`cli.resolveDeps`, `projects.ResolveDep`). The project part is a link, else a registered project, which is linked in the same command (the link change is applied first, `withLinks`). A suffix resolves in that project's checkout; a project not checked out here needs the full ID. Dropped targets are refused. Stored as `<link>:<full ID>` in `depends_on`. On every load `Tree.External` says what each is here: found (state, title, checkout), missing (checked out, no such issue: X005 warning), unknown (not checked out: shown with the clone hint, never blocking, never reported). A found one that is not done blocks like a local dependency; one dropped there is X006. `prep show` lists them (`cross_deps` in JSON, also in list summaries), `prep guide` points at found ones, and `prep show <id> --dependents` scans other checkouts for issues depending on this one ("blocks in <project>"); "blocks" stays within the project otherwise. Cycles across projects are not detected. In the domain (`links.go`): `Apply` renames references in memory (`RenameLinkInText`) so `CheckWrite` sees a rename's result and carries `External`; `DepState` makes `UnresolvedDeps` skip unknown references, and `PlanNew`/`PlanEdit` refuse dropped targets through it.
