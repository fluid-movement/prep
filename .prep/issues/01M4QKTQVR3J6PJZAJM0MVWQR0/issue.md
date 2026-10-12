---
title: 'Agent behavior test suite: prompts run against fake projects, sessions analyzed'
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - testing
  - adoption
  - integration
priority: high
---

Reliable tool use by agents is only provable by running agents. This suite builds fake prep projects, runs a Claude Code session on a prompt in one of them, and checks that the expected tools were used and the expected outcome came to be. Deterministic wherever possible: a test runner that knows the prompt, the fixture project and the other projects it needs; assertions on the normalized transcript and on the project state afterwards. An agent judge only for what cannot be checked mechanically. Results are pass rates over repeated runs, so the same suite can compare variants of hooks, skills and prompts (mutation testing) and later other harnesses. Claude Code only at first; other harnesses are a later issue. Ties into the reliability and tool-use issues: hooks, ledger analysis, activity history, classification.

Related reliability and tool-use issues this suite measures and backs: 01M4J8789E81QAJHG8R8ZPEMRD (Claude Code hooks make sure prep calls are made), 01M4J87R4AFB193QF0G3AECNA0 (hooks in the popular harnesses), 01M4J8784Q2JR9V8EWAH0C0A3F (can Jev classify whether a tool call should be a prep call), 01M4J8JVQ7TVGMEF0BB2AZRFG8 (activity stream keeps enough history to measure agent behaviour), 01M4J8780KZRKWHJR4NN8AD85M (ledger analysis, done).

## Open questions
