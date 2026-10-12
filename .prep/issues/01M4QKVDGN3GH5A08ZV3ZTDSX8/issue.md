---
title: 'Agent behavior tests on other harnesses: adapters for Codex, Gemini, Cursor and others'
kind: code
parent: 01M498VQD4Q85CZYK965WW3H9Q
depends_on:
  - 01M4QKTQVR3J6PJZAJM0MVWQR0
tags:
  - testing
  - integration
---

The test suite starts with Claude Code only. This adds a transcript adapter and runner per harness, so the same cases check that each harness's hooks and skills make agents use prep well.

## Open questions
