---
title: Better preview of themes
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - theming
---

Choosing a theme should not need guesswork. Today the gallery cycles themes one at a time with `t`/`T`, and the settings screen switches and saves immediately. A better preview lets a user compare themes and see one on their own issues before committing to it.

## Open questions

- What should the preview show: the real issue screen with the project's data, the gallery, or a compact sample (a few list rows, a detail, badges) per theme?
- Side by side (several themes at once, compact) or one at a time with a theme list to move through?
- Should the settings screen preview while moving through themes and save only on enter, with esc restoring the previous theme?
