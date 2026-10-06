---
title: Homebrew tap for macOS and Linux
kind: code
depends_on:
  - 20261005-152624
tags:
  - install
---

Many macOS users expect brew install. GoReleaser publishes a formula to a tap repository (fluid-movement/homebrew-tap) on every release, so brew install fluid-movement/tap/prep and brew upgrade work; prep update already defers to brew for Homebrew installs.

## Open questions

- Is a separate tap repository acceptable, or should the formula live elsewhere?
