---
title: Restructure user docs into docs/ with a generated command reference
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
  - docs
---

Implement the docs decision "User docs are Markdown pages in docs/, command reference generated" (`prep knowledge find user docs`).

Create `docs/` with pages: getting-started, workflow (the lifecycle: define, ready, claim, complete, checks), commands (generated), harness-integrations (Claude Code, Codex, `prep setup`), cloud-sessions (moved from the README), knowledge-base, tui. Move the matching README content into the pages and shrink the README to the pitch, install, a quick start and a table of links to the pages; nothing is lost, only moved.

The command reference `docs/commands.md` is generated from the command table that `prep help` prints. Add `prep docs commands` (or a `just` recipe that runs a small Go program; choose the lighter one) that writes it, and make `just ci` fail when the committed file differs from the generated one. Links between pages are relative and checked by a test or recipe in `just ci`. The cloud page quotes `prep cloud script` by reference, never a copy.

## Open questions
