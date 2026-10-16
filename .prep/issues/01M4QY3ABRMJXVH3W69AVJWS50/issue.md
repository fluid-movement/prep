---
title: 'Bound the output agents read back: list, check and help'
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQSEP2EKTYKYFC79EGFS6
  - 01M52ZRB3VY0MSZ0HZWJ2SEC24
tags:
  - cli
  - tokens
---

In the real sessions of the other prep projects the biggest part of what prep prints is `prep list` (21 calls, 81,000 characters, 3,900 per call), `prep check` (up to 26,000 characters in one call), `prep help` (4 calls, 7,000 each) and `prep guide` (2,400); `prep knowledge find` returns 8-9 KB per query (2026-10-16). Everything an agent reads is paid again on every later request. After the removals of the rethink release have shrunk these commands, bound the outputs by default and say how to see more, without losing information agents need. Absorbs 01M52ZRZZGEJGY10AEJHHR2CHN (knowledge find).

## Open questions
