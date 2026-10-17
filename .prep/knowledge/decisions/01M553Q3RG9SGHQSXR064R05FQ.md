---
type: decision
title: Skill carries rules.md verbatim, kept in sync by a test
description: skill.md is a hand-kept short intro plus rules.md, and a test fails when it drifts.
date: "2026-10-17"
issue: 01M52ZRM0R2QPKK4HZE6PV62CE
---

A go generate step adds a build stage for a 700-byte file. A test that skill.md ends with the Workflow rules heading and rules.md text keeps one source (rules.md, embedded for prime) with no tooling; just sync-skill still copies skill.md to the three plugin copies.
