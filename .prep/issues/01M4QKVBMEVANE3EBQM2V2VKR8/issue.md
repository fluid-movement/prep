---
title: Transcript capture and normalization to a harness-neutral format
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
tags:
  - testing
---

## Requirement

A Claude Code session is captured and turned into a harness-neutral transcript that every assertion and the judge read. The capture is the `stream-json` output of a headless run; the normalized form has: session metadata (model, harness version, plugins, skills and tools visible), the tool calls in order (name, input, result, error flag), prep commands parsed out of Bash calls into command, arguments and target project, the assistant's text, the final result, and counts (tool calls, errors, tokens in and out, cost, duration).

Only the adapter knows Claude's format; the normalized type does not. The raw stream is kept beside the normalized transcript so a run can be re-analysed without running the agent again.

## Open questions
