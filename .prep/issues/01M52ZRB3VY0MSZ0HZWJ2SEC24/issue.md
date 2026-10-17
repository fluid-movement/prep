---
title: Knowledge as decision and finding records
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQZBAS48ZK64MW7H85QJ4
  - 01M52ZR294ZMVPKF7N2KPPKD9Q
tags:
  - knowledge
priority: high
---

prep decide writes a decision record into the knowledge base, linked to its issue or standalone, with optional supersedes and code-path scopes; prep finding writes a finding record (replacing findings.md). Records are never edited; superseded status is derived. Remove knowledge new, update, confirm, bootstrap (per the bootstrap decision) and review, drift and confirmation handling, and the knowledge completion gate. Decision and research issues complete when a linked record exists. Keep OK Format; resolve index.md merge conflicts by regeneration through a git merge driver. Decided in 01M4ZCM5XQ3J6TM1N4AZ330ZH7 and 01M4ZCMFJ5P67XNSMKX1NPANDF; trial to revisit after a month.

Settled with the user, 2026-10-17:
- One file per record, named by ULID: knowledge/decisions/<ulid>.md and knowledge/findings/<ulid>.md. The generated index.md per directory is resolved by the merge driver.
- An issue's decisions.md goes away: prep decide <id> writes a knowledge record linked to the issue, and prep show <id> lists the linked records by title. Existing issue decisions are converted by 01M52ZRE24EMGBHX3SKV62436G.
- --supersedes <record id> works across the knowledge base, replacing per-issue D<n> numbers; superseded status is derived and shown on the older record.
- --scope <path glob> is repeatable and optional; a record without scope is project-wide. Here scopes are stored and searchable; surfacing them on touched paths belongs to the finish hook issue (01M52ZRSZFBMWZQJPAXVHXD7M1).
- Every record has a required one-line description; prep knowledge find and show stay the read path.

## Open questions
