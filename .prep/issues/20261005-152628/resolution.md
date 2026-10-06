---
outcome: done
by: claude-code/2.1.289
at: 2026-10-06T06:43:54Z
documentation:
  no_impact: the storage format entry already documents this format; tolerance of blank lines lands with 20261006-064318
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---
