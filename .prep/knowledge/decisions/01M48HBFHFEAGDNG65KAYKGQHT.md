---
type: decision
title: "ULIDs only, no migration: convert this repository once"
description: "ULIDs only, no migration: convert this repository once"
date: "2026-10-06"
issue: 01M485SBR0WN8JDQSN7A3GT7CB
supersedes:
  - 01M48HBFHENMRMZFNR5738F78S
---

The user's call: prep is unreleased and this repository is its only user, so prep ships no migration. The schema version stays 1; `ValidID` accepts only ULIDs, and timestamp directories become I001 errors. This repository is converted once by a throwaway script, run from the scratch area and not committed. For each issue it builds a ULID from the issue's timestamp (the time part) and random bits, renames the directory, and rewrites every occurrence of the old ID in `.prep` text files. Afterwards `prep check` is clean. Baseline file names stay timestamps: they are per issue, not IDs.

Alternatives considered: a schema 2 step in `prep migrate` (the previous D3), which is code kept forever for a single conversion; keeping both formats (D2), which the user ruled out.
