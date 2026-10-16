---
title: Eval records harness errors as errors, not failed assertions
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - eval
priority: high
---

When a session ends with a harness error (for example 'You've hit your session limit', result is_error true), prep eval records the run as an error, excludes it from pass rates and says so in the report. Today such runs count as completed and failed: 57 of 100 runs in 20261016-101957-23ca were scored this way. Decided in 01M4ZCMHFA382AG6Q8T9MXFVQT. Comes first, because every later behavior change is measured with evals.

## Open questions
