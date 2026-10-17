---
type: decision
title: Releases are tag-driven and versions must agree
description: Pushing a v* tag releases; the workflow first checks every plugin/marketplace manifest carries the tag's version.
date: "2026-10-17"
scope:
  - .github/workflows
  - justfile
---

Pushing a v* tag releases; the workflow first checks every plugin/marketplace manifest carries the tag's version.

(implied) marketplace pinned to the release tag (= batch 1 in part).
