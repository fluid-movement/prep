---
type: decision
title: Guide is for solo work; delegate implementer and checker passes by brief and handback
description: After the approved three-issue trial, hand-offs use brief files and prep-free reports; guide remains the complete workflow for an agent working alone.
date: "2026-11-11"
issue: 01M74V9A3E6Y9N3KMS9QGHCY1J
supersedes:
  - 01M529ANGJXESXDB191W6XQK2F
---

Adopts the existing hand-off decision 01M72CVCP45FN7D515257JH6VE after the approved trial recorded in finding 01M7648TW8B9QK8M953A8BC85P. This replaces the older guidance to prompt delegated agents to run prep guide. The trial held the file/report format on three real issues with a small model and zero prep calls by implementers and checkers.

For a ready issue, prep guide <id> remains the complete contract for an agent working alone: requirement, criteria, context, applicable decisions, implementation and separate checking/completion instructions. It also tells an orchestrator how to hand off instead.

The orchestrator runs prep brief <id> --out and gives only the one-line prompt "Read .prep/local/handoff/<id>/brief.md and follow it exactly. It tells you where to write your report." It imports the report with prep handback <id>, generates prep brief <id> --check --out for a separate checker, and points the same prompt at check-brief.md. Before prep handback <id> --check it verifies knowledge updates and other tracker-only Definition of Done items itself. Implementers and checkers follow their brief, make no prep calls and need not receive the workflow rules. Knowledge-lookup agents still receive the rules. Parallel implementers keep separate worktrees.

Reason: the orchestrator is the single record writer and the brief replaces handwritten protocol prompts without conflating implementation and checking. The user's external orchestrate skill needs the same change, but is not edited in this repository. The new eval case is added but left for the user to run live.
