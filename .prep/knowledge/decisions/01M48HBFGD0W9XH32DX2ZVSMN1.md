---
type: decision
title: "Evidence is informational; validate the format only"
description: "Evidence is informational; validate the format only"
date: "2026-10-06"
issue: 01M46ARQ78P8M40K4FSPVNK531
outcome: true
---

The hash names the work as it was done; after a squash merge the same change lives in another commit, and the record should not turn into a warning. Validating the format keeps typos out. Alternatives: warn about unreachable hashes (noise after every squash merge); require reachable commits (forces completing after the merge and breaks branch workflows).
