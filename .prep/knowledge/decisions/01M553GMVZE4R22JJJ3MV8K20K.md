---
type: decision
title: Prompt context is global, small by default
description: Config hooks.prompt off/small/large; small is default; large adds guides at a cost; invalid config makes it silent; hooks never edit the file.
date: "2026-10-17"
scope:
  - internal/cli/hook.go
  - internal/userconfig/**
---

Large pays for issue loading and guide context (stated); small keeps cost down.
