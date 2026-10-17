---
type: decision
title: "prep link add <project> [--as name]"
description: "prep link add <project> [--as name]"
date: "2026-10-11"
issue: 01M4N5KBDT4SMJ8EBG33MKZKN9
---

The requirement wrote `prep link add <name> <project>` with the name defaulting to the project's own name. An optional leading positional is awkward, so the project comes first and the local name is `--as`, defaulting to the target's `name:`.
