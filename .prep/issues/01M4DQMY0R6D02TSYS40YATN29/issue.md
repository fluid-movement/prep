---
title: Pi package with the prep skill
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMFXQZN7JK8SD2Z55ZTDH
tags:
  - integration
  - pi
---

A Pi package in `plugins/pi` carries the prep skill, so a Pi agent finds out how to use prep the way a Claude Code agent does.

- The skill is `internal/cli/skill.md` unchanged: `just sync-skill` copies it into the package and a test checks the copy, as for the Claude Code plugin.
- The skill and the `prep prime` briefing read correctly in Pi as well as Claude Code: Pi has no built-in subagents, so the advice to hand surveys and knowledge lookups to subagents applies where the harness has them, and agents without them keep output bounded inline.
- The package follows Pi's package practices: a `pi` manifest naming its extension and skill, host packages as `*` peer dependencies, no build step.
- The extension sets `PREP_ACTOR` to `pi/<version>` for the session, so records name the agent without relying on the model to pass `--by`.
- Loading the package from the checkout during development is documented, and the skill was seen loading in a live Pi session.

## Open questions
