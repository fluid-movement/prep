---
type: decision
title: JSON output contract and exit codes
description: Under --json all output, errors included, is JSON on stdout; exit codes are 0 ok, 1 gate/check failure, 2 usage, 3 conflict.
date: "2026-10-17"
scope:
  - internal/cli/**
---

Under --json all output, errors included, is JSON on stdout; exit codes are 0 ok, 1 gate/check failure, 2 usage, 3 conflict.

Reason not stated (agents parse it).
