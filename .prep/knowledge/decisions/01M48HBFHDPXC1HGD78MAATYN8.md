---
type: decision
title: "Keep timestamp IDs valid; order by creation time"
description: "Keep timestamp IDs valid; order by creation time"
date: "2026-10-06"
issue: 01M485SBR0WN8JDQSN7A3GT7CB
---

Both formats are valid issue IDs. A helper returns an ID's creation time: the timestamp for the old format, the 48-bit time for a ULID. Every place that orders issues by ID (`IDs`, query tie-breaks, tree order) compares creation times, then the ID string. This keeps existing directories, references and commit messages unchanged (the user's choice of the recommended option).

Alternative considered: migrating every issue to a ULID once. One format, but every directory, every parent and dependency reference, and every ID in past commit messages and notes would change or dangle.
