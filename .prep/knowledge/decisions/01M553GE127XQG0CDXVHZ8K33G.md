---
type: decision
title: Marketplace pinned to the binary's release tag
description: Install adds the marketplace at the binary's tag; both plugin manifests carry the same version and the release refuses a tag that differs.
date: "2026-10-17"
scope:
  - internal/setup/**
  - .claude-plugin/**
  - plugins/**
---

Third-party marketplaces do not auto-update (stated).
