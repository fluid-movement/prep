---
type: decision
title: "The git remote is a project's identity; names are handles mapped onto remotes"
description: "The git remote is a project's identity; names are handles mapped onto remotes"
date: "2026-10-11"
issue: 01M4N5ARVE6J5APZYJ7VA0N464
---

The user proposed the git repository as the project's ID, so a cross-project link resolves on every machine that has the other project checked out. Records use short names (`api:01M4…`) instead of URLs so they stay readable; each project's `project.md` maps the names it uses to remotes (`links:`), and its own `name:` is initialized from the repository name. Remotes are normalized to `host/owner/repo`. A project that is not checked out resolves to "unknown", never blocking. Rejected: directory basenames as identity (differ per clone, collide).
