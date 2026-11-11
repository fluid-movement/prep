---
type: decision
title: Brief leaves out prep-command DoD items, bounds decisions at 6,000 bytes, and takes --out with an optional value
description: prep brief drops Definition of Done items that name a prep command, fills the decisions section greedily within 6,000 bytes (own cut, others dropped and counted), and resolves --out's optional value by position.
date: "2026-11-11"
issue: 01M72CVD0W8E28HQZVRA9FD92K
scope:
  - internal/domain/brief.go
  - internal/cli/brief.go
---

Choices made while building `prep brief`:

- Definition of Done items that name a prep command ("prep check passes", "prep fmt --check passes") are left out of a brief. An agent without prep cannot run them and must not touch the records, and a brief names no prep command; the orchestrator runs those checks at hand-back. Rejected: printing them, which breaks the prep-free rule for every project with the default Definition of Done.
- The decisions section is bounded at 6,000 bytes, 200 of them reserved for its notes. Decisions are taken in order: the issue's own, ancestors' nearest first, then scoped ones by the literal length of the matching scope. An own decision that does not fit is cut when at least 200 bytes of body remain, and marked; any other that does not fit is dropped and counted in the section. Later, smaller ones may still fit.
- `--out` takes an optional value by position, because Go's flag package has none: `--out=<file>` always names the file; `--out <arg>` does when another non-flag argument remains to be the issue; otherwise the default place under `.prep/local/handoff/<id>/` is used.
- Running `prep brief` creates the hand-off directory even when it prints to standard output, so the agent can write the report with any tool.
- Only prep's own wording is free of prep commands; decision texts, context and Definition of Done items are project content and may mention them.

Seen on real issues: scoped decisions fill the bound when the context names broad files (26 listed by title and 19 left out for the brief issue itself).
