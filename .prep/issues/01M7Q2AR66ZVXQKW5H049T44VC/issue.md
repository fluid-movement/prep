---
title: 'prep handoff: hand ready issues to cloud sessions'
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
depends_on:
  - 01M7Q2AQV2EXF21424KDFS7CNY
tags:
  - cloud
---

`prep handoff <id>...` hands ready issues to Claude Code cloud sessions and brings the results back. It is the command form of what the cloud spike did by hand (findings "Cloud spike: one issue from push to merge" and "Cloud spike, part 2", `prep knowledge find cloud spike`). Two subcommands, nothing else:

**`prep handoff start <id>...`**
- For each issue it checks what `prep brief` already enforces (code or research issue, ready, not stale, dependencies done, unclaimed) and that `HEAD` is pushed to its upstream, because a cloud session clones from GitHub. It refuses, naming what is missing, and starts nothing when any issue fails.
- It starts one `claude --cloud` session per issue. `claude --cloud` refuses to run without an interactive terminal, so prep runs it on a pseudo-terminal and reads the session id and link it prints; when prep itself has no terminal it prints the exact command for the user to run instead and exits non-zero.
- The prompt is the standard one: run `prep guide <id>`, implement it, log "Implemented, ready for check", do not complete, commit the work and the records, push the branch, and write the final report as the issue's handback (`prep handback` format) into the records on the branch, because a session's chat answer does not come back with its branch.
- It records each started session in the issue's work log with `prep log` (session id and link) and prints one line per issue. It never pushes, so the user pushes first; it prints how many sessions it is about to start and asks to confirm unless `--yes`.
- The environment is whatever `/remote-env` chose; prep does not select it.

**`prep handoff fetch`**
- Lists the `claude/*` branches on origin that are not merged into the current branch and, for each, the issue it touches and whether its report is present.
- With `--merge` it merges them into the current branch one at a time (never one octopus merge: parallel branches conflict in generated index files), runs `prep fmt` and then `prep check` after each merge, stops at the first failure and reports it. It never pushes.

Out of scope: the checker pass (run it as a local or second cloud pass with the existing checker flow), polling or watching sessions, choosing the model, and usage limits (prep cannot see them; the orchestrator's usage check stays in the skill).

## Open questions
