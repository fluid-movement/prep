---
title: Activity stream keeps enough history to measure agent behaviour
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - activity
priority: high
---

The activity stream is capped at 1 MiB and pruned to its newest 2,000 lines, which at this repository's rate (800 events on a busy day) is two to three days. Release 0.3.0 measures how often prep is used where expected before and after each hook change, and the ledger analysis (01M4J8780KZRKWHJR4NN8AD85M) had to fall back to Claude Code's own transcripts for anything older than 2026-10-08. The stream also lost its earlier history when an agent deleted the file during a test.

Keep the history the measurements need without slowing the hot path: for example, rotating pruned lines into dated archive files under `.prep/local` that `prep activity` can read with a time range, with a size or age limit set in the config.

## Open questions

- How long must history be kept, and is a size limit or an age limit the better knob?
