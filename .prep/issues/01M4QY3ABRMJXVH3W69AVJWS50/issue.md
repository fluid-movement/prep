---
title: 'Bound the output agents read back: list, check and help'
kind: code
parent: 01M4QY2WM6XN0JSSTHR5TGQFJT
tags:
  - cli
  - tokens
---

In the real sessions of the other prep projects the biggest part of what prep prints is `prep list` (21 calls, 81,000 characters, 3,900 per call), `prep check` (up to 26,000 characters in one call), `prep help` (4 calls, 7,000 each) and `prep guide` (2,400). Everything an agent reads is paid again on every later request, and the project's principle is token efficiency. Bound these outputs by default and say how to see more, without losing information agents need.

## Open questions
