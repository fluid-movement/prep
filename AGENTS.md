# AGENTS.md

prep is a workflow engine and memory for coding agents, with a human supervising. It is a Go CLI and TUI that keeps issues, their lifecycle records and a knowledge base of decisions and findings in the repository under `.prep/`, versioned with the code. This repository is prep itself.

## Where things live

- `cmd/prep`: the binary's main package.
- `internal/domain`: the core model, lifecycle rules and ports; no storage knowledge.
- `internal/mdstore`: the markdown storage adapter for `.prep/`.
- `internal/okf`: the OKF knowledge bundle (entries, index generation).
- `internal/cli`: commands and output; `internal/cli/skill.md` is the source of the prep skill.
- `internal/tui`: the terminal UI; `internal/palette`: its colors and themes.
- `internal/activity`: per-machine agent activity stream; `internal/watch`: file watching.
- `internal/gitx`: git access; `internal/projects`: project registry.
- `internal/setup`: harness integration install (`prep setup`); `internal/update`: self-update; `internal/userconfig`: `~/.config/prep`.
- `internal/eval`, `eval/`: evaluation of prep with agents.
- `plugins/`: the Claude Code and Codex plugins.
- `install.sh`, `justfile`, `README.md`, `TOKEN-EFFICIENCY.md`.

## Build, test, check

- `go test ./...` (or `just test`) runs all tests.
- `just install` builds with the version stamped, installs the binary and refreshes the harness integrations from this checkout.
- `just lint` runs gofmt and go vet; `just fmt` formats Go and the prep records.
- `prep check` validates the records; `prep fmt --check` verifies canonical form. `just check` runs both from source; `just ci` runs lint, test and check.
- After changing `internal/cli/skill.md`, run `just sync-skill` so the plugin copies match.

## Work and decisions

- Work on this repository is tracked with prep itself. Start an issue with `prep guide <id>`; `prep prime` gives the project briefing.
- Change prep records only through prep commands, never by editing their files.
- Decisions and findings live in the knowledge base. Find them with `prep knowledge find <words>` and read them with `prep knowledge show <entry>`; record new decisions there, not in code comments or docs files.
- Commit straight to `main` unless told otherwise.
