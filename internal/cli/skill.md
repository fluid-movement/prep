---
name: prep
description: Work on issues tracked by prep (.prep/ in the repository). Use when the user mentions an issue ID, asks what to work on next, or asks to define, enrich, implement or complete an issue.
---

# prep

This repository tracks its work with prep. Issues live in `.prep/issues/<id>/`; the knowledge base lives in `.prep/knowledge/`.

1. At session start, run `prep prime` for a short briefing. If it says the knowledge base is not bootstrapped, tell the user and work on the bootstrap first: knowledge entries can only be written after `/overview.md` exists.
2. For the issue you work on, run `prep guide <id>`. It tells you the current step, what to read, which gates are unmet, where outputs go, and the exact command for the next transition. Follow it; it always matches the installed binary.
3. Use `--json` when you parse output. Reads (`prime`, `guide`, `list`, `next`, `show`, `check`) never change anything; writes (`new`, `edit`, `context`, `decide`, `criterion`, `dod`, `findings`, `log`, `knowledge`, `import`, `define`, `ack`, `ready`, `claim`, `release`, `complete`, `drop`, `fmt`, `fix`, `migrate`) each perform one change.
4. Write through commands, not files. `prep edit <id>` changes title, kind, parent, dependencies, tags, priority or requirement (`--depends-on` replaces the whole list, `--depends-on ''` clears it; `--priority critical|high|medium|low`, medium being the default; `--body`/`--body-file` replaces the requirement including its `## Open questions` section). `prep next` and `prep list` order work by priority, then ID. Records: `prep context` and `prep findings` replace their text, `prep decide` appends a decision, `prep criterion` adds and checks criteria by the numbers `prep show` prints, `prep dod` changes Definition of Done additions and opt-outs, `prep log` appends to the work log. Knowledge entries: `prep knowledge new <entry> --type --title --description --body-file -`, `prep knowledge update <entry>` (only the fields you pass), `prep knowledge confirm <entry>` or `--drifted` after re-checking entries against their scoped code. Long text goes through `--body-file -` (stdin).
5. Never invent issue IDs; `prep new` creates them (ULIDs: 26 characters, sortable by creation time). Commands accept any unique suffix of an ID, such as the last 6 characters the TUI shows. Never write `baselines/`, `ready.md`, `claim.md` or `resolution.md` by hand.
6. When the user wants to bring in work tracked elsewhere (TODO comments, GitHub issues, other trackers), run `prep import`: it creates one research issue whose guide walks through finding the sources, reviewing candidates with the user, and creating each as an open issue with a `Source:` line. `prep list --text` matches titles and requirements, so it finds an issue by its source.
7. Pass `--by <harness>/<version>` (or set `PREP_ACTOR`) so records name the agent.

## Without the binary

If `prep` is not on the PATH (for example in a cloud session), install it first; it takes under a minute:

```sh
curl -fsSL https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"
```

`| PREP_VERSION=v0.1.0 sh` picks a version. Without network access to GitHub releases, `go install github.com/fluid-movement/prep/cmd/prep@latest` works wherever Go is installed (the binary lands in `$(go env GOPATH)/bin`). Then use the commands above as usual.

Only when both fail are the files the database and you edit them directly. Tell the user that prep could not be installed and that you are editing `.prep` by hand; `prep check` (later, locally or in CI) validates the edits:

- Lifecycle: open (`issue.md`) → defined (`baselines/<timestamp>.md`) → ready (`ready.md`, references the newest baseline) → in progress (`claim.md`) → done or dropped (`resolution.md`). State is derived from which records exist.
- Define: the requirement is the prose in `issue.md`; an issue can only be defined when its `## Open questions` section is empty.
- Enrich: `context.md` (implementation context; link knowledge entries as `/path.md`), `decisions.md` (append-only `## D<n>: <title>` entries with `date:` and optional `supersedes:` / `outcome: true`), `acceptance.md` (`- [ ]` criteria, optional `## Definition of Done` additions and `- opt-out: <item> — <reason>`).
- Complete: all criteria checked; code issues need a commit; research issues need `findings.md`; decision issues need an `outcome: true` decision; every completion records which knowledge entries changed or why there was no impact.
- Do not edit record files by hand; leave the transition for when the binary is available, and say so.
