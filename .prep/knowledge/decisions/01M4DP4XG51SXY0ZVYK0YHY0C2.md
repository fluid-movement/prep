---
type: decision
title: "One copy of the inference and usage rules"
description: "One copy of the inference and usage rules"
date: "2026-10-08"
issue: 01M46ARM9G3492V298TM82KD45
---

Issue inference (`infer.ts`) exists twice already (Claude Code panel, token-ledger) and usage areas once (`classify.ts`). A Pi git install clones the whole repository, so the Pi extension imports one shared, dependency-free module directly, and the Claude Code copies, which must stay inside their plugin directories, are checked identical to it by a test, as the skill copies are. The Pi extension follows bash and powershell results and edit and write paths, including nested calls (`parentToolCallId`), and only successful ones.
