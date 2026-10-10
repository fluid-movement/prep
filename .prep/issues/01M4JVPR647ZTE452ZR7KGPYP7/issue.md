---
title: How agents verify the desktop app's UI
kind: research
parent: 01M4JVP89F3PGDHYETHEPS3YBB
tags:
  - desktop
---

Agents build most of prep, and a native UI is harder for them to check than the TUI (program tests) or a web page (a browser). Find a workflow that lets an agent see and verify its UI changes: SwiftUI snapshot tests, dumping the accessibility tree, screenshots of a running build, XCUITest from the command line. Findings pick one and say what the scaffold needs for it.

## Open questions
