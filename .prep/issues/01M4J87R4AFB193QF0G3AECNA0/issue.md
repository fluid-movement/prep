---
title: Hooks in the popular harnesses make sure prep calls are made
kind: code
parent: 01M498VQD4Q85CZYK965WW3H9Q
depends_on:
  - 01M4J8789E81QAJHG8R8ZPEMRD
tags:
  - integration
  - adoption
---

Bring the hooks proven in Claude Code (01M4J8789E81QAJHG8R8ZPEMRD) to the other popular agent harnesses, so an agent in any of them uses prep where it should and its activity reaches `prep tui`. Each harness gets the same behaviour through its own hook or extension mechanism, installed and removed by `prep setup`; where a harness lacks a hook point, its child records what it does instead. Each child is one harness.

## Open questions
