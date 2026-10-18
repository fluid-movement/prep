---
title: The release builds for Windows
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
priority: high
---

The v0.3.0 release failed: GoReleaser builds windows targets, and internal/eval (env/probe.go, judge/exec.go, run/run.go) used the Unix-only Setpgid and syscall.Kill. CI did not catch it because it builds only for Linux.

## Open questions
