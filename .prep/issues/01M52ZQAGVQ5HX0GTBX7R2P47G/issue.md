---
title: Remove the completion contract machinery
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQDGE4WPA096TW2EHMZ4N
tags:
  - workflow
priority: high
---

Remove the completion contract machinery, per the completion decision (01M4ZCM21C7D5S14442QW1TT7F): prep policy (set, show, adopt, retain) and the completion_policy and completion_policy_history fields in project.md; prep evidence and verification.md; prep exception and approvals.md; contract hashes and bindings in ready.md and resolution.md (contract, previous_contracts, verification, approvals, dod and dod_opt_outs snapshots); issue- and ancestor-level Definition of Done with opt-outs and their approval (prep dod approve); knowledge review reports (prep knowledge review, knowledge-reviews.md), which are bound to contracts; the scoped worktree fingerprints (internal/worktree) and the I029 warning and G_POLICY and G_EVIDENCE gates.

Keep one project-level Definition of Done, shown as guidance in the briefing and in prep guide at the check step, not a gate. prep dod --add and --remove edit the project Definition of Done (today prep dod only works on issues, and agents cannot edit project.md by hand). Completion requires checked criteria and the documentation decision (--docs or --no-impact), which the knowledge issue (01M52ZRB3VY0MSZ0HZWJ2SEC24) removes later.

This repository's records and the other prep projects on this machine are converted once with a contract stage in prep-wt/convert-records.py, run with the user's go-ahead after the new binary is installed; no compatibility code. Delete the code, its tests, fixtures, eval cases and knowledge text together. Done when nothing references the removed parts and GitHub CI on main is green.

## Open questions

- User's OK of this requirement (drafted 2026-10-17, agreed in direction: conversion option a, a script the user lets run).
