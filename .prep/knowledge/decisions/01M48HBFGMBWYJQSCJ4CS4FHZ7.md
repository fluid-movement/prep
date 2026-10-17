---
type: decision
title: "Pin the marketplace to the binary tag"
description: "Pin the marketplace to the binary tag"
date: "2026-10-06"
issue: 01M46P009GD5Y6ACBYRMN00D4V
---

Third-party marketplaces do not auto-update, and an unpinned marketplace follows main, so the plugin could run ahead of or behind the binary. Pinning the marketplace to the binary's tag makes the installed plugin exactly match the installed binary; setup moves the pin when the binary changes. Alternative: follow main with a compatibility range (weaker guarantee, more checks).
