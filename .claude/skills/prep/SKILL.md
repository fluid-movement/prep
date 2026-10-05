---
name: prep
description: Work on issues tracked by prep (.project/ in the repository). Use when the user mentions an issue ID, asks what to work on next, or asks to define, enrich, implement or complete an issue.
---

# prep

This repository tracks its work with prep. Issues live in `.project/issues/<id>/`; the knowledge base lives in `.project/knowledge/`.

1. At session start, run `prep prime` for a short briefing.
2. For the issue you work on, run `prep guide <id>`. It tells you the current step, what to read, which gates are unmet, where outputs go, and the exact command for the next transition. Follow it; it always matches the installed binary.
3. Use `--json` when you parse output. Reads (`prime`, `guide`, `list`, `next`, `show`, `check`) never change anything; writes (`new`, `define`, `ack`, `ready`, `claim`, `release`, `complete`, `drop`, `fmt`, `fix`, `migrate`) each perform one transition.
4. Never invent issue IDs; `prep new` creates them. Never write `baselines/`, `ready.md`, `claim.md` or `resolution.md` by hand.
5. Pass `--by <harness>/<version>` (or set `PREP_ACTOR`) so records name the agent.

## Without the binary

If `prep` is not installed (for example in a cloud session), the files are the database:

- Lifecycle: open (`issue.md`) → defined (`baselines/<timestamp>.md`) → ready (`ready.md`, references the newest baseline) → in progress (`claim.md`) → done or dropped (`resolution.md`). State is derived from which records exist.
- Define: the requirement is the prose in `issue.md`; an issue can only be defined when its `## Open questions` section is empty.
- Enrich: `context.md` (implementation context; link knowledge entries as `/path.md`), `decisions.md` (append-only `## D<n>: <title>` entries with `date:` and optional `supersedes:` / `outcome: true`), `acceptance.md` (`- [ ]` criteria, optional `## Definition of Done` additions and `- opt-out: <item> — <reason>`).
- Complete: all criteria checked; code issues need a commit; research issues need `findings.md`; decision issues need an `outcome: true` decision; every completion records which knowledge entries changed or why there was no impact.
- Do not edit record files by hand; leave the transition for when the binary is available, and say so.
