---
type: decision
title: Bridge sets PREP_SESSION and PREP_ACTOR
description: The Claude bridge sets session and actor env at session start and on session ID change so Bash prep commands and bridge events count as one agent.
date: "2026-10-17"
scope:
  - plugins/claude-code/hooks/**
---

So reads are recorded without --by and commands and events meet as one agent (stated).
