---
type: component
title: OKF store
description: internal/okf — immutable OKF decision/finding writes, reads and generated indexes.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/okf
confirmed_commit: 28f150511800c7e93ff7bcd5e1ed4f499c1f8651
---

# OKF store

Reads `.prep/knowledge` as an OKF v0.2 bundle. Existing descriptive entries remain readable. Markdown files except index.md and log.md are parsed for frontmatter and body; invalid frontmatter becomes a diagnostic. Links resolve relative to the entry, or from bundle-root paths. Code fences and spans do not contribute links.

Records live in decisions/<ulid>.md or findings/<ulid>.md. ParseEntry reads type, title, description, date, optional issue, scopes, supersedes IDs and decision outcome. The filename supplies the ID. The domain derives superseded status from newer records. There is no git drift or confirmation handling.

Render builds a new KnowledgeRecord and parses it back for CheckWrite. Apply syncs a temporary file then atomically links it into its destination, refusing any existing file, including one created after Render or unreadable as an entry. No command edits a record. Required fields are type, title, one-line description and date; issue, scopes, decision supersedes and outcome are optional. Body holds rationale or a finding.

IndexFiles generates index.md per bundle directory, including descriptions and root okf_version 0.2. WriteIndexes atomically updates generated indexes after writes and in fmt/fix; K007 reports missing or outdated indexes. Init writes a union merge rule for indexes, which fmt/fix regenerate after a merge. See [Knowledge conventions](/conventions/knowledge-base.md).
