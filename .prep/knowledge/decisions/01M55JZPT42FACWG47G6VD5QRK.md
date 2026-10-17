---
type: decision
title: Look up prior decisions and capture user decisions with their reason
description: Before changing code, look up and follow existing decisions; record decisions the user states with their reason, without recording small implementation choices.
date: "2026-10-17"
issue: 01M55JY43XT800M50Y6YFC5YJ0
scope:
  - internal/cli/rules.md
  - internal/cli/skill.md
  - internal/cli/hook_finish.go
---

The user specified these guidance changes in this issue after eval run A showed agents recording their own narrow choices instead of the user decision and failing to look up existing decisions before coding. The workflow rules carry both requirements through the briefing and skill; the shared finish reminder asks about decisions made or received, including ones the user stated.
