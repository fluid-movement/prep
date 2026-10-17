---
title: Eval fixtures retry a prep start that hits text file busy
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - eval
---

TestLiveRunner failed in CI on 2026-10-17 with 'fork/exec .../bin/prep: text file busy' (ETXTBSY): on Linux a parallel run still copying its pinned binary can leak the write descriptor into a fork. The fixture's prep exec helpers retry briefly on ETXTBSY; the process never started, so a retry is safe.

## Open questions
