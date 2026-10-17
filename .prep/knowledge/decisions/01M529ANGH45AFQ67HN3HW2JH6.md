---
type: decision
title: "Hooks are the main guidance channel"
description: "Hooks are the main guidance channel"
date: "2026-10-16"
issue: 01M4ZCM7TTPASZ2EN8MQN6EPVZ
---

User decision, 2026-10-16, backed by the plugin A/B (hooks only 12/13, skill only 3/11, none 0/9). Hooks that call prep commands are the main way agents get guidance: deterministic, and prep crafts the right text for each step. In a project without prep the hooks return nothing (verified for all five hooks). Downside: the same context is injected repeatedly; to be solved by injecting less or only what changed (open question). The skill is no longer the main carrier and stops being a command manual; commands live in prep guide and --help, which match the binary.
