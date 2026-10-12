---
title: Baseline run of the agent behavior suite, and file what it finds
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVD1SBR79P1QYECT9KYQ9
tags:
  - testing
---

Run the initial suite for real (just eval-full, with CLAUDE_CODE_OAUTH_TOKEN from claude setup-token): first prep eval doctor and prep eval judge-check, then the suite on the current variant with sonnet, N=5. Keep a small baseline of the smoke subset's pass rates in eval/baseline.json. File every failure that reveals a gap in the skill, hooks or CLI as an issue under the Claude Code hooks issue 01M4J8789E81QAJHG8R8ZPEMRD. Needs a Claude token, so the user runs it or provides the token.

## Open questions
