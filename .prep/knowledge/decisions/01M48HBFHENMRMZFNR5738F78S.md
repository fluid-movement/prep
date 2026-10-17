---
type: decision
title: "ULIDs only, from schema 2; prep migrate converts existing projects"
description: "ULIDs only, from schema 2; prep migrate converts existing projects"
date: "2026-10-06"
issue: 01M485SBR0WN8JDQSN7A3GT7CB
supersedes:
  - 01M48HBFHDPXC1HGD78MAATYN8
---

The user ruled out mixed formats: prep is in development and this repository is its only user. Schema 2 accepts only ULIDs as issue IDs. The migration from schema 1 builds each issue's ULID from its timestamp (the 48-bit time part, random bits from the injectable source), so creation order is kept. It renames the issue directories and rewrites every occurrence of an old ID in `.prep` text files: frontmatter references, records, bodies, logs and knowledge entries. A project on schema 1 gets P002 until it is migrated. Baseline file names stay timestamps: they are per issue, not IDs.

Alternatives considered: keeping both formats valid (the previous D2), which the user ruled out; rewriting only frontmatter references, which would leave dangling IDs in requirements, contexts and logs.
