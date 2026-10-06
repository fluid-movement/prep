---
title: Refuse empty bodies in knowledge writes
kind: code
tags:
  - knowledge
  - cli
---

prep knowledge update --body-file with an empty file wiped an entry's body without complaint, which happened while documenting the release work. An empty body is refused with a usage error for both new and update, so a failed script cannot silently erase an entry.

## Open questions
