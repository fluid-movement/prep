---
title: Write commands for knowledge entries
kind: code
depends_on:
  - 20261005-184730
---

The CLI is prep's API, and the knowledge base sits behind its own store port (the OKF adapter today), so agents should create and update knowledge entries through commands rather than writing files under .prep/knowledge. Commands carry the entry content and metadata (type, title, description, status, scope, confirmed commit) and validate before writing: links resolve, the entry is not renamed, required frontmatter is present. prep guide lists them in the documentation step.

## Open questions

- Command shape: one prep knowledge subcommand group (new, update, confirm) or top-level commands?
- Should confirming an entry against the current commit (setting confirmed_commit) be its own command?

## Open questions
