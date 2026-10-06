---
title: Cloud sessions without the prep binary
kind: research
tags:
  - adoption
---

Cloud agent sessions (Claude Code on the web and similar) start from a fresh checkout without the prep binary. They fall back to editing the files directly, guided by the skill, so every rule the binary enforces (gates, validation, canonical form, record files only commands write) depends on the agent following prose. As write commands replace file editing, the gap between local and cloud sessions grows. Find out how cloud sessions can run the real binary, or what the fallback must guarantee when they cannot.

## Open questions

- Which harnesses matter first (Claude Code web, GitHub Actions agents, others), and what can each install at session start?
- Is a release binary fetched by the SessionStart hook enough, or is a fallback needed at all (for example a check in CI that rejects invalid hand edits)?
