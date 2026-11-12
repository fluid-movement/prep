---
type: decision
title: 'Derived scope is not built: broad directory scopes already fill the notice cap'
description: 'Measured on prep''s history: 42% of linked unscoped decisions gain a derived scope, but explicit scopes already match 16-25 decisions per typical file, so a five-decision cap hides derived ones'
date: "2026-11-12"
issue: 01M74VAE370N8SQXGKRVG2B9WM
---

Finding 01M77A84CW51YRW6FF6JGNT83J: of 171 linked, unscoped, current decisions, 72 (42%) gain a derived scope (median 10.5 files, max 17). But the explicit scopes alone match 16 to 25 decisions per typical file in internal/cli and internal/tui, because broad directory scopes (internal/tui on 21 decisions, internal/cli on 12, internal/cli/hook*.go on 9) crowd one another. The requirement's stop rule applies: with the cap of five, derived decisions would never reach the notice, so the derived-scope work is not built.

Reason: building it would add a history cache and ranking rules that change nothing the agent sees. The effective fix is narrower explicit scopes, which the stop question and prep decide now ask for (01M72CX0AH92JGYWED5YM5EMW6).
How to apply: do not reopen derived scope unless the explicit scopes have been narrowed first; re-measure before building.
