---
title: Cloud sessions without the prep binary
kind: research
parent: 01M48FAKAG2X998F2ZHE1MHVTC
tags:
  - adoption
---

Cloud agent sessions (Claude Code on the web and similar) start from a fresh checkout without the prep binary. They fall back to editing the files directly, guided by the skill, so every rule the binary enforces (gates, validation, canonical form, record files only commands write) depends on the agent following prose. As write commands replace file editing, the gap between local and cloud sessions grows.

Research Claude Code on the web first: what a session can install at start (for example a release binary fetched by the SessionStart hook), whether that is enough, and what a fallback must guarantee when the binary cannot run (for example a CI check that rejects invalid hand edits). Other harnesses follow once that path is clear.

## Open questions
