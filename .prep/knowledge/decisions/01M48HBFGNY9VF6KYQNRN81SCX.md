---
type: decision
title: "Prime does the version check"
description: "Prime does the version check"
date: "2026-10-06"
issue: 01M46P009GD5Y6ACBYRMN00D4V
---

The hook passes the plugin root to prep prime, which reads plugin.json and compares versions. The comparison lives in Go, tested, and the hook stays one line. Alternative: compare in a shell script inside the plugin (untested, duplicated per harness).
