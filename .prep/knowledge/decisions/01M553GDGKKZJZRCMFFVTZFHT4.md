---
type: decision
title: One skill source, copies kept identical
description: internal/cli/skill.md is the source; the Claude plugin skill and .claude/skills/prep/SKILL.md must stay identical (just sync-skill, tested). Codex copy is generated from it plus Codex notes. (merges cli.md, codex.md)
date: "2026-10-17"
scope:
  - internal/cli/skill.md
  - plugins/*/skills/**
  - .claude/skills/**
---

internal/cli/skill.md is the source; the Claude plugin skill and .claude/skills/prep/SKILL.md must stay identical (just sync-skill, tested). Codex copy is generated from it plus Codex notes. (merges cli.md, codex.md)

Reason not stated.
