---
type: decision
title: Restrict title-stack escapes to terminal files
description: Only push/pop the xterm title stack when the output is an `*os.File` attached to a terminal.
date: "2026-11-11"
issue: 01M5FS4DPG9B02CXB7PMTZQVA8
---

This avoids escape output for pipes and test writers; the pop is deferred so it also runs when the TUI exits with an error.
