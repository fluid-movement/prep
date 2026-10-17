---
type: decision
title: Direct .prep access is redirected to the CLI
description: before-tool denies supported reads and writes of prep records with CLI guidance; it is a guardrail, not enforcement.
date: "2026-10-17"
scope:
  - internal/cli/hook*.go
---

(implied) records must go through checked writes
