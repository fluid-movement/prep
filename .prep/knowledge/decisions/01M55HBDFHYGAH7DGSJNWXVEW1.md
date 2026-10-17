---
type: decision
title: Bound only the lists that grow with the project
description: Default text reads hide resolved issues, list the 30 newest knowledge records and cap ranked search at five blocks and 3,000 characters, with recovery flags.
date: "2026-10-17"
issue: 01M4QY3ABRMJXVH3W69AVJWS50
scope:
  - internal/cli/**
  - internal/domain/query.go
supersedes:
  - 01M554P2YP7JSQDKHJ37N2M84F
---

Measured after the rethink removals: knowledge list was 56 KB and list 33 KB, while check was one line and help 4 KB. Bound the growing lists and leave help and guide unchanged. Default issue list hides done and dropped issues; explicit state filters and saved views retain their behavior, and --state all recovers resolved issues. Knowledge list shows 30 newest records, active before superseded, with --max N or --all to recover the rest. Knowledge find retains block ranking and grouping, selecting at most five blocks (therefore at most five entries) and 3,000 Unicode characters including the truncation marker and footer. An explicit --max N bypasses the character budget, and --max 0 shows every matching block; footers count partially printed blocks as omitted. JSON lists and default JSON search stay complete. Parent show hides resolved children only when the child list exceeds 30 lines, naming prep list --under ID --state all to recover them. Live scale and session evals require user approval.
