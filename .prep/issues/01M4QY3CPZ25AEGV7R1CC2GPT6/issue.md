---
title: The project registry forgets directories that no longer exist
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - cli
  - projects
priority: low
---

Any prep command run in a directory records it in the machine-wide registry ([Project registry](/components/projects.md)), and nothing removes an entry when its directory is deleted: scratch projects under the system temp directory and removed git worktrees stay in prep projects and in every project's briefing ('Other prep projects here'), where agents see them and may try to use them. When prep reads the registry it drops entries whose root no longer exists and saves the cleaned registry. Worktrees are now the standard way to hand off issues (01M52ZR8564ECW8GP3ZHSXEM0A), so this matters more. Decided in 01M4ZCMDMWBYMYDMBNQKBYG1JH.

## Open questions
