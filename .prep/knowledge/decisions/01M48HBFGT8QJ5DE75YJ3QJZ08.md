---
type: decision
title: "Check runs on demand with drift"
description: "Check runs on demand with drift"
date: "2026-10-06"
issue: 01M46QA9T0FJG0DQ5VNKARFPWE
---

The check screen needs knowledge drift, which calls git and costs about 170 ms per load. Running it only while the screen is open keeps live reload of the issue views at about 10 ms. Alternative: always load with drift (slower reloads for a screen rarely open), or never show drift in the TUI (hides warnings prep check reports).
