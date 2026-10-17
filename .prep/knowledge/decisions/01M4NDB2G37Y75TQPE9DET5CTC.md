---
type: decision
title: "The CLI computes identity; mdstore stays free of git"
description: "The CLI computes identity; mdstore stays free of git"
date: "2026-10-11"
issue: 01M4N5BKPZ1JAPSZC27P1272B4
---

Deriving the default name needs origin's URL. mdstore has never called git, so the CLI (through `internal/projects`) computes the name and a new ULID and hands them to `Init` and `SetIdentity`. The registry reads identity through `mdstore`'s parser, so there is one reader of project.md.
