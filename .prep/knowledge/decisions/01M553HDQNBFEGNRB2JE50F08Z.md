---
type: decision
title: Go, one static binary
description: Core, CLI and TUI ship as one static Go binary.
date: "2026-10-17"
scope:
  - go.mod
  - cmd
---

Near-instant startup for frequent agent calls, simple install; TypeScript needs a runtime and starts slower, Rust has no decisive advantage and slower iteration.
