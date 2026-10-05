---
title: Write commands for enrichment and work records
kind: code
depends_on:
  - 20261005-184730
---

The CLI is prep's API: every change to the data goes through a command that carries content, and the storage adapter decides the representation. Agents today write context, decisions, acceptance criteria, history and findings by editing files, and prep guide lists file paths under Write. Each of these records gets a write command with --json and validation before writing, prep guide lists those commands instead of paths, and the skill tells agents to use them. Hand edits of the markdown stay possible for humans and remain caught by prep check.

## Open questions

- Granular commands that match each record's semantics (decisions append-only, criteria added and checked one by one, history appended) or one generic write per record with the content from a file or stdin?
- Does the skill's "Without the binary" path for cloud sessions keep file editing as the fallback?
