---
title: 'Test isolation: own HOME and config, pinned prep binary, no leakage'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVB5EW2N6Y9BDVCFJDC7Q
tags:
  - testing
---

A run must not see the real user's hooks, skills, memory or projects registry. The fixture gets its own HOME and Claude config, the prep binary under test is pinned, and nothing touches the network or the real repo.

## Open questions
