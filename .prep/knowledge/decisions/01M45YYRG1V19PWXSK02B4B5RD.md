---
type: decision
title: "Project-local configuration, no plugin yet"
description: "Project-local configuration, no plugin yet"
date: "2026-10-05"
issue: 01M46ARJB0BYKFHTN456ZAS5SB
---

The hook and commands live in this repository's `.claude/` directory. A Claude Code plugin through a marketplace is the intended distribution for other projects, but it has to check binary compatibility, which depends on the install and update flow (01M46ARN8RW78DATMQV6P9NVYK). Alternatives: a plugin now (premature without a release channel); `prep init` writing `.claude/` files (couples the core to one harness).
