---
title: Pi package with the prep skill
kind: code
parent: 01M46ARM9G3492V298TM82KD45
tags:
  - integration
  - pi
---

A Pi package carries the prep skill, so a Pi agent finds out how to use prep the way a Claude Code agent does. It is the base the briefing, pane and setup issues build on.

- **Package**: `plugins/pi` holds the extension and the skill; a private root `package.json` (keyword `pi-package`, a `pi` manifest) points at them, so the repository installs as `git:github.com/fluid-movement/prep@vX.Y.Z` (D4). Host packages are `*` peer dependencies, nothing is bundled, and there is no build step. The extension factory only registers handlers; nothing starts until `session_start`.
- **Skill**: `internal/cli/skill.md` unchanged; `just sync-skill` copies it into the package and a test checks the copy, as for the Claude Code plugin. No project skill copy for Pi: it would collide by name with the package skill.
- **Wording**: the skill and the `prep prime` briefing read correctly in Pi as well as in Claude Code. Pi has no built-in subagents, so handing surveys and knowledge lookups to subagents applies where the harness has them; without them the agent reads narrowly and keeps output bounded inline.
- **Actor**: the extension sets `PREP_ACTOR` to `pi/<pi version>` for the session, so records name the agent without the model passing `--by`.
- **Shared rules**: issue inference and usage areas move to one dependency-free shared module that the Pi extension imports; the Claude Code panel's and token-ledger's copies are checked identical to it by a test (D5).
- **Tests**: Node's test runner with type stripping for pure parts, and an SDK session with pi-ai's faux provider for the extension's wiring (D6), run by a `just` recipe next to the Claude Code plugin tests.
- Loading the package from the checkout during development is documented, and the skill was seen listed and loaded in a live Pi session.

## Open questions
