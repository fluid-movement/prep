---
title: 'Desktop app scaffold: Swift package, build, CI, distribution'
kind: code
parent: 01M4JVP89F3PGDHYETHEPS3YBB
depends_on:
  - 01M4JVPQXB93KH02V3WJS1HF0C
tags:
  - desktop
---

A SwiftPM-based macOS app in the repo (e.g. `desktop/`) that finds the `prep` binary, opens a project, lists its issues through `prep list --json` and refreshes from `prep watch`. `just` targets build and test it; a macOS CI job runs them. Release builds are signed and notarized and ship as a Homebrew cask next to the formula.

## Open questions
