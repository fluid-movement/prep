---
title: Desktop app (macOS)
kind: manual
tags:
  - desktop
---

A native macOS app as a second human UI next to the TUI: the TUI's feature set, plus switching between projects, and later launching coding agents from the app. Linux users keep the TUI; Windows is out of scope.

The app is Swift (SwiftUI, AppKit where SwiftUI is weak for keyboard-driven dense lists). It talks to prep only through the CLI: `prep … --json --root <dir>` for reads and writes and one `prep watch` subprocess per open project for refresh. It never reads `.prep/` directly, so there is one domain implementation and the app exercises the same contract agents use. Its children are the work that ships the first version; agent launching is v2.

## Open questions
