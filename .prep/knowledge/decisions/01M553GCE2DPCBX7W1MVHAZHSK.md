---
type: decision
title: Domain owns lifecycle rules
description: CLI and TUI share one domain model and transition planners; storage adapters load and persist but never implement lifecycle rules.
date: "2026-10-17"
scope:
  - internal/domain/**
  - internal/mdstore/**
  - internal/okf/**
---

CLI and TUI share one domain model and transition planners; storage adapters load and persist but never implement lifecycle rules.

So both interfaces agree on derived states and transitions (stated).
