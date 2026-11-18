---
type: finding
title: Cloud handoff commands carry reports in versioned work logs
description: handoff start validates all issues and pushed HEAD, starts PTY sessions and logs links; fetch merges sequentially and formats/checks after each.
date: "2026-11-18"
issue: 01M7Q2AR66ZVXQKW5H049T44VC
---

prep handoff start accepts ready, unclaimed code and research issues, reusing BriefGate for kind, staleness, parent and dependency gates. It verifies the live upstream and requires an interactive terminal, otherwise printing exact claude --cloud commands. The shared cloud prompt stores the full handback-format report under Cloud handback: in the issue work log with prep log, then commits and pushes the code and records. This preserves research reports without importing them through prep handback, which would complete research automatically. prep handoff fetch lists unmerged origin claude/* branches and report presence, optionally merging one at a time followed by in-process fmt and check. Neither local command pushes; the checker pass stays separate.
