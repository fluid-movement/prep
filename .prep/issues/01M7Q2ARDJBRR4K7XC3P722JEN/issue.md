---
title: Decide and start proper user docs
kind: decision
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

Decide how prep's user documentation is organised, then create its skeleton. The README carries install, workflow, harness setup and soon cloud use; it should shrink to an overview with links.

Settle: a `docs/` folder of Markdown pages (getting started, the workflow, commands, harness integrations, cloud sessions, the knowledge base) versus a generated site, and what stays in the README. User-facing docs say how to use prep; decisions and findings stay in the knowledge base, as AGENTS.md requires. Decide how the command reference stays in step with `prep --help`, so it is generated or checked rather than written twice, and how `just ci` keeps links and examples valid.

Deliver the decision as a knowledge entry and a skeleton of pages with the existing README content moved into them.

## Open questions
