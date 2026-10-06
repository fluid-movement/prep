---
title: Handle issue IDs that collide after a merge
kind: code
parent: 20261006-112434
tags:
  - storage
---

Issue IDs are second-resolution timestamps, so two branches can create issues with the same ID; when they merge, both issues' files land in the same directory and git either reports conflicts or silently mixes two issues. prep check recognizes a directory that holds more than one issue (conflicting or duplicated records) as a guided diagnostic, explains the situation, and prep offers a way to split the second issue into a new ID without losing either issue's records.

## Open questions

- Which signals are reliable after a merge: git conflict markers, two different titles in history, mixed baselines?
- Should prep new reduce the risk up front, for example by adding a short random suffix to IDs, or keep plain timestamps?
