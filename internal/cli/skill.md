---
name: prep
description: Work on issues tracked by prep (.prep/ in the repository). Use when the user mentions an issue ID, asks what to work on next, or asks to define, enrich, implement or complete an issue.
---

# prep

This repository tracks its work with prep: issues and knowledge live in `.prep/`. If no session hook did, run `prep prime` at session start. Run `prep guide <id>` for the issue you work on, and `prep --help` for commands.

Workflow rules:
- Settle the requirement with the user before you define an issue.
- Ready means another agent without this conversation could do the work.
- Implementers claim, implement exactly the issue, log "Implemented, ready for check" and stop; they never complete, commit or push.
- Checkers verify each criterion and the Definition of Done themselves, say which commands they ran, then complete.
- Before changing code, look for decisions about it with prep knowledge find <area>; follow them, or record a new decision with prep decide --supersedes <id> first.
- Record a decision with prep decide when it is made, including one the user states, with the reason given; your own small implementation choices need no record.
- Read and write records only through prep commands, never the files; search knowledge with prep knowledge find and read it with prep knowledge show.
- Give parallel implementers separate git worktrees.
- Run live evals only with the user's OK.
- Keep command output short, and never read back a saved output file.
- Give knowledge lookups across several entries to a subagent on a small model, and pass these rules on to every subagent.
