---
type: decision
title: 'Hand-off report format: headed sections, numbered criteria verdicts, ### blocks for decisions and findings'
description: 'One definition in internal/domain/report.go for brief and handback: sections Summary, Criteria, Decisions, Findings, Problems; criteria lines ''<n>. met — evidence''; decisions and findings as ''### title'' blocks with an optional Scope line.'
date: "2026-11-11"
issue: 01M72CVD0W8E28HQZVRA9FD92K
scope:
  - internal/domain/report.go
---

Defined in `internal/domain/report.go` for `prep brief` and `prep handback`, so the two cannot drift.

- Sections are `## Summary`, `## Criteria`, `## Decisions`, `## Findings`, `## Problems`; empty ones may be left out. A checker reports Summary, Criteria and Problems; a research implementer reports Summary, Findings, Decisions and Problems.
- A criteria line is `<number>. met — <evidence>` or `<number>. not met — <what is missing>`.
- A decision or finding is a line `### <title>` followed by its text, with an optional `Scope: <glob>` line directly under the title. Reason: a reason of several lines parses without ambiguity, which a single title-and-reason line does not allow.
