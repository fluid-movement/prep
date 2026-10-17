---
type: decision
title: Screens never set colors or styles
description: Screens compose ui components; colors come from theme tokens, never literals at a use site.
date: "2026-10-17"
scope:
  - internal/tui
---

Screens compose ui components; colors come from theme tokens, never literals at a use site.

Calm, consistent look; change colors only in internal/palette/builtin.go.
