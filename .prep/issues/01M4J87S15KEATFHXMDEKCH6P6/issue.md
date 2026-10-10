---
title: 'Pi: hooks make sure prep calls are made'
kind: code
parent: 01M4J87R4AFB193QF0G3AECNA0
tags:
  - integration
---

Port the Claude Code prep hooks (01M4J8789E81QAJHG8R8ZPEMRD) to Pi through its own hook or extension mechanism: the briefing at session start, the redirects of misdirected calls under `.prep`, and the session's tool events sent to `prep activity add`. `prep setup` installs, refreshes and removes them. A failure in a hook never changes or blocks a tool call.

## Open questions

- Which hook points does Pi offer, and which of the Claude Code hooks can it carry?
