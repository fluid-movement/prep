---
type: decision
title: Writes are planned, then checked, then applied
description: Every write is a domain.Change; Tree.CheckWrite validates it in memory and rejects writes that add errors before the adapter renders and writes.
date: "2026-10-17"
scope:
  - internal/domain/**
  - internal/cli/**
---

Every write is a domain.Change; Tree.CheckWrite validates it in memory and rejects writes that add errors before the adapter renders and writes.

Reason not stated.
