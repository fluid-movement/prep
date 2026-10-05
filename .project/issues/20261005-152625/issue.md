---
title: Completion evidence after squash merges
kind: decision
---

A branch commit hash recorded as completion evidence disappears after a squash merge. Decide whether unreachable hashes are accepted as informational or only reachable ones are verified. Today prep validates the hash format only; knowledge drift reports unknown confirmed commits as warnings.

## Open questions
