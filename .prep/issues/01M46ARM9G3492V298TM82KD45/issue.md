---
title: Pi setup
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
depends_on:
  - 01M48721G8Y21HT0ZGAXB5QZF2
tags:
  - integration
  - pi
---

prep works in Pi as well as it works in Claude Code: a Pi agent gets the same skill and briefing, and its activity (prep commands, tool calls, tokens) reaches the Agent view of `prep tui`, which the person runs next to Pi. `prep setup` installs and updates the integration like the Claude Code plugin. The integration is a Pi package served from this repository: the skill and one thin extension over the CLI, with no FSM logic ([Stack](/decisions/stack.md)).

- the prep skill, identical to `internal/cli/skill.md` and worded so it holds in either harness
- the briefing from `prep prime` as a system prompt section, with the version warning, and the session's tool and request events sent to `prep activity add`
- installation, refresh and removal through `prep setup`, pinned to the binary's version

The extension follows Pi's practices: nothing starts in its factory, long-lived resources start at `session_start` and stop idempotently at `session_shutdown`, it works in print, JSON and RPC modes, the briefing keeps the prompt prefix stable for caching, and a failure in the extension never changes a tool call.

What Pi can do beyond this (own tools, prompt shaping) comes as separate issues.

## Open questions
