---
type: decision
title: Harness failures are never failed assertions
description: Errors (including is_error with subtype success) get error status and are excluded from pass rates, shown as an errored count.
date: "2026-10-17"
scope:
  - internal/eval/run/**
  - internal/eval/matrix/**
---

Errors (including is_error with subtype success) get error status and are excluded from pass rates, shown as an errored count.

(implied) do not blame the agent for harness failure
