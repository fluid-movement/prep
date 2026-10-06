# Handoff: finish the open work of Release 0.1.0

You are a cloud agent working on prep, a workflow engine and memory for coding agents. This repository tracks its own work with prep (`.prep/`). Three issues of Release 0.1.0 are ready and enriched; your job is to implement them one after another, in the order below, and leave the repository green after each.

## Set up prep first

The cloud session starts without the `prep` binary, but this repository is prep itself, so build it from source (Go 1.26 or newer; `go.mod` says `go 1.26.0`):

```sh
go build -o "$HOME/.local/bin/prep" ./cmd/prep   # or any directory on PATH
export PATH="$HOME/.local/bin:$PATH"
export PREP_ACTOR="claude-code/cloud"            # every write records who made it
prep prime                                       # the briefing: what is open, what is actionable
```

Rebuild the binary whenever your own changes touch `internal/` and you want to use the new behavior (`go build` again). If Go is missing or the build fails, say so in the session result and stop: do not edit `.prep` files by hand.

Read `.claude/skills/prep/SKILL.md` (or `prep skill`) once; it is the contract for working with prep. In short:

- `prep guide <id>` tells you the current step, what to read, the unmet gates and the exact next command. Follow it.
- Write through commands (`prep claim`, `prep log`, `prep criterion --check n`, `prep findings`, `prep complete`, `prep knowledge update|confirm`), never by editing record files.
- The Definition of Done for every issue: `go test ./...` passes, `prep check` reports no errors, `prep fmt --check` passes, and knowledge entries describing changed behavior are updated (`prep knowledge update <entry> --body-file -`, then `prep knowledge confirm <entry>` after the commit that changed the code, so drift warnings clear).

## The work, in order

### 1. Cloud sessions without the prep binary — `20261005-185825` (research)

Do this first: you are running in exactly the environment it asks about. Record what this session offers at start (OS, architecture, network access to github.com, whether plugin hooks and setup scripts run, writable PATH directories) as evidence, then recommend how a cloud session should get the binary and what a fallback must guarantee. Write the result with `prep findings <id> --body-file -`. Create follow-up issues for the recommended changes with `prep new` (kind code, tag `adoption`, **no parent**: the user schedules them into a release), link them from the findings, and complete the issue with `--no-impact` or the entries you changed.

### 2. Collision-free issue IDs (ULID) — `20261006-083752` (code)

ULIDs replace timestamp IDs completely: prep is unreleased and this repository is its only user, so there is no compatibility to keep and no migration to ship. `prep new` creates ULIDs, ULIDs are the only valid issue ID (the schema version does not change), and the TUI shows a ULID's last 6 characters. The context and the decisions (D1 and D4; D4 supersedes D2 and D3) say where IDs are generated, validated, ordered and displayed. Keep tests deterministic by injecting the random source. Do this before item 3, because item 3 creates issues with `prep new`.

Finish it by converting this repository once, with a throwaway script you run from a temporary directory and do not commit: give each issue a ULID whose time part is its old timestamp, rename its directory, and replace every old ID in `.prep` text files. Rebuild the binary, check that `prep check` is clean, and commit the conversion on its own. **After that, every issue ID in this file is stale.** Find the remaining work by title with `prep list --text "Import existing work"` or `prep next`.

### 3. Import existing work items — `20261006-063354` before the conversion (code)

A `prep import` command that plans one guided research issue, modeled on the knowledge bootstrap (`internal/domain/bootstrap.go`). The decision on the issue lists the steps its context must guide. Sources are project-specific on purpose: no per-source importers.

## Per issue

1. `prep guide <id>`, read what it lists, `prep claim <id>`.
2. Implement; keep commits focused. Match the surrounding code: its comment density, naming and idioms. Tests sit next to the code; the TUI has golden snapshots (`go test ./internal/tui/... -update` only for intended changes, and say why in `prep log`).
3. Check criteria off with `prep criterion <id> --check n` as they become true; note anything a reviewer should know with `prep log <id> "<text>"`.
4. Run the Definition of Done checks, commit, then `prep complete <id> --commit <sha> --docs <entry>...` (or `--no-impact "<reason>"`), confirm the touched knowledge entries, and commit the records.

Commit messages: a short imperative subject, a body that says what changed and why, and this trailer:

```
Co-Authored-By: Claude <noreply@anthropic.com>
```

The project convention is to commit straight to `main`. If the session can only push its own branch, work there and say so in the result.

## Do not

- Publish anything: no tags, no releases, no edits to the release workflow's version numbers. Release 0.1.0 itself (`20261006-112434`) is the user's: its remaining criteria are manual.
- Touch Release 0.2.0 or its children (`20261006-123443`).
- Change a requirement. If an issue's requirement turns out wrong or unclear, leave it, add a `prep log` line explaining the problem, and move on to the next item.
- Edit `.prep` files by hand or invent issue IDs.

## When you finish or get stuck

End with a short report in the session result: what was completed (issue IDs and commits), which follow-up issues you created, anything you skipped and why, and the state of `prep check` and `go test ./...`.
