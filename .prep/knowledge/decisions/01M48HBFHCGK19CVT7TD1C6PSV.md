---
type: decision
title: "Generate ULIDs with the standard library, entropy injectable"
description: "Generate ULIDs with the standard library, entropy injectable"
date: "2026-10-06"
issue: 01M485SBR0WN8JDQSN7A3GT7CB
---

A ULID is 48 bits of milliseconds and 80 bits from `crypto/rand`, encoded in Crockford base32. That is a few dozen lines, so prep implements it in `internal/domain` instead of adding a dependency (Stack: few dependencies). Within one tree, a new ULID must sort after every existing ID created in the same millisecond, so generation retries or increments like `nextStamp` does today. The random source is injectable so tests stay deterministic.

Alternatives considered: `github.com/oklog/ulid` (a dependency for code this small); a timestamp with a random suffix (readable, but the user chose ULIDs).
