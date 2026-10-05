---
title: Decide the install and update flow
kind: decision
---

Decide how prep is installed and updated. Likely: GoReleaser on tag push, GitHub Release binaries, a curl | sh installer, and prep update self-replacing (deferring to Homebrew or go install when installed that way). The Claude Code plugin ships separately through a plugin marketplace and checks binary compatibility. It must be ergonomic and rock solid.

## Open questions
