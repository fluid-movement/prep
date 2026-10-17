---
type: decision
title: Project identity is the ULID, not name or remote
description: A project is its permanent ID in project.md; name is a handle and the remote only says where to clone.
date: "2026-10-17"
issue: 01M4N5ARVE6J5APZYJ7VA0N464
scope:
  - internal/projects
  - internal/mdstore
---

Renames, remote changes and moves keep identity; a fork stays the same project until `prep init --new-id` (01M4N5ARVE6J5APZYJ7VA0N464 D2).
