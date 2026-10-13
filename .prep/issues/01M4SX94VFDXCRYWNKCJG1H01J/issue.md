---
title: Configurable prompt hooks control injected Prep context
kind: code
tags:
  - integration
  - hooks
  - tokens
---

Add a shared Prep hook entry point for harness prompt hooks, with global configuration controlling how much context it injects. When prompt hooks are off, the hook command should return immediately with no Prep lookup. When enabled, it should first detect whether the current working directory is a Prep project and return immediately otherwise. Support context levels: a low-cost mode that injects a concise workflow reminder, and a richer mode that parses the prompt for a Prep issue ID and includes the relevant issue guide/context when found. Keep harness-specific adapters thin and document the token and runtime tradeoffs.

## Open questions
- What should the configuration key and default level be?
- Should the levels be named off/small/large, or use a different scale?
- Which prompt hook events and harnesses can support the shared behavior, and what fallback should be used where prompt hooks are unavailable?
- Should issue-ID parsing accept only full IDs, or unique suffixes too?
