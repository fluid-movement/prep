---
type: decision
title: "Record layout: one file per transition, no empty files, 0.3.0 records removed"
description: "Record layout: one file per transition, no empty files, 0.3.0 records removed"
date: "2026-10-16"
issue: 01M4ZCMFJ5P67XNSMKX1NPANDF
---

User agreed, 2026-10-16. Keep one directory per issue with state derived from which transition files exist. Create a record file only when something is written to it (today 145 of 221 decisions.md and 63 history.md are empty). Remove the 0.3.0 record types (verification, approvals, knowledge-reviews, resolution-reviews, resolution-history, dependency-overrides, contract fields) and the unused artifacts frontmatter (zero uses). decisions.md and findings.md move to knowledge records. Minimal reopen moves the old resolution into history.md. context.md and acceptance.md stay separate files so a reader loads only what it needs. history.md gets merge=union in .gitattributes so parallel appends on branches do not conflict.
