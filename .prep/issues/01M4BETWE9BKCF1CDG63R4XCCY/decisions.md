## D1: Activate on prep init and on the panel commands
date: 2026-10-08

An inactive panel retries `prep prime` after the agent's successful `prep init` and when `/prep:pane` or `/prep:focus` runs. The first covers the reported case without the person doing anything; the second covers `prep init` run outside the agent.

Alternatives: retrying on every turn or every Bash call while inactive costs a `prep prime` process per call in every non-prep session; watching for `.prep` to appear needs a file watcher for a directory that may never exist.
