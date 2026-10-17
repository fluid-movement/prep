---
type: decision
title: Issue IDs are ULIDs, resolved by unique suffix
description: IDs are ULIDs (same-millisecond IDs increment to keep order) and any unique suffix, either case, resolves.
date: "2026-10-17"
issue: 01M485SBR0WN8JDQSN7A3GT7CB
scope:
  - internal/domain/ulid.go
  - internal/domain/tree.go
---

String order equals creation order (stated); ID format decided in 01M485SBR0WN8JDQSN7A3GT7CB, reason there.
