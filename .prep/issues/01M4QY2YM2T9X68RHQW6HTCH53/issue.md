---
title: 'Plugin A/B: what do the skill and the hook each buy?'
kind: research
parent: 01M4QY2WM6XN0JSSTHR5TGQFJT
tags:
  - adoption
  - testing
  - tokens
priority: high
---

Measure what each part of the Claude Code plugin changes in agent behavior: the skill (`plugins/claude-code/skills/prep/SKILL.md`, about 3,000 tokens in context for every session that loads it) and the SessionStart hook with the activity bridge (`hooks/hooks.json`, `hooks/register.ts`, `prep prime --hook`). Four variants: `none` (no plugin at all), `skill-only`, `hook-only`, `current`. Answer: does the behavior the plugin is meant to produce (finding and using prep, tracking work, reading knowledge through prep) come from the skill, from the hook briefing, from both, or from neither (the agent finds prep anyway)? Done when the findings state, per behavior, which part carries it and what each part costs in tokens.

## Open questions
