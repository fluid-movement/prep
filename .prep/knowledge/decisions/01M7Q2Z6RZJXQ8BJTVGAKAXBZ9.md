---
type: decision
title: User docs are Markdown pages in docs/, command reference generated
description: docs/ pages, README as overview, command reference generated from prep help and checked in CI; no site yet.
date: "2026-11-18"
issue: 01M7Q2ARDJBRR4K7XC3P722JEN
---

The user delegated the choice. Markdown in the repository is readable on GitHub and by agents, needs no build or hosting, and versions with the code. A site adds hosting and a toolchain without a reader who needs it yet; the same pages can feed one later. The command reference is the part that drifts, so it is generated from `prep help` and checked in just ci. Decisions and findings stay in the knowledge base; docs only say how to use prep. Alternative considered: everything in the README, which already carries too much.
