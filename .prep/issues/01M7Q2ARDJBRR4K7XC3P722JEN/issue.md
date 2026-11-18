---
title: Decide and start proper user docs
kind: decision
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

Decide how prep's user documentation is organised. The README carries install, workflow, harness setup and now cloud use; it should shrink to an overview with links.

Settled: user docs are Markdown pages in `docs/` in the repository, readable on GitHub and by agents, with no generated site for now. The README keeps the pitch, install, a quick start and links. Pages: getting started, the workflow (lifecycle), commands, harness integrations, cloud sessions, the knowledge base, the TUI. The command reference is generated from `prep help` by a command and checked in `just ci`, so it cannot drift. User docs say how to use prep; decisions and findings stay in the knowledge base as AGENTS.md requires. A site can be added later over the same pages.

Deliverable of this issue: the decision recorded in the knowledge base. The restructuring is a separate code issue.

## Open questions
