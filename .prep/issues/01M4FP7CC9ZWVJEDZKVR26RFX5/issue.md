---
title: One grammar for both navigation tiers; no crumb, no agent dot
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - design
---

The header still looked inconsistent: the screens bar marked the current screen with a bold accent underline, the views with a background chip and │ separators, the Agent item carried a status dot, and a `‹` crumb appeared and disappeared beside `prep`.

Both tiers now share one grammar (`navStyle`): plain labels, the current one accent and bold, the others muted, spaced from the scale. The crumb goes (backspace still goes back), and the bar shows no agent dot; the Agent screen's second tier shows its liveliness.

## Open questions
