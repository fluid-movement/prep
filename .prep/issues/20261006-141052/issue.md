---
title: 'Skill: install the binary before falling back to hand edits'
kind: code
tags:
  - adoption
---

The skill's "Without the binary" section tells an agent to edit `.prep` files directly as soon as `prep` is not on the PATH. The research in 20261005-185825 showed that a cloud session can usually get the binary in under a minute: the release installer (`install.sh` with `PREP_VERSION`) downloads from GitHub, and `go install github.com/fluid-movement/prep/cmd/prep@<version>` works through proxy.golang.org wherever Go is installed.

Change the section (in `internal/cli/skill.md`, the plugin copy and this repository's `.claude/skills/prep`) so an agent first tries to install the binary with those two commands, and only hand-edits when both fail. In that case it says so to the user, keeps to the rules it already lists, and expects `prep check` (locally later or in CI) to validate the edits.

## Open questions
