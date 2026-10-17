---
type: decision
title: "No schema bump for the optional field"
description: "No schema bump for the optional field"
date: "2026-10-06"
issue: 01M4874AR04JFPZ5KYW0402JC1
---

priority is an optional, additive frontmatter field like tags, which also arrived within schema 1; absent means medium, so every existing file stays valid. An older binary rejects only issue.md files that set a priority (KnownFields), and prep update resolves that. Alternative: schema 2 with a no-op migration, which would make older binaries refuse the whole project instead of the few prioritized issues.
