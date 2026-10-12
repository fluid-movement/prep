---
title: Transcript capture and normalization to a harness-neutral format
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
tags:
  - testing
---

Capture a Claude Code session (stream-json or the transcript file) and normalize it: tool calls in order, prep commands parsed into command and args, errors, final message, token and call counts. Assertions are written against this format so other harnesses only need an adapter. Consider reusing the activity stream.

## Open questions
