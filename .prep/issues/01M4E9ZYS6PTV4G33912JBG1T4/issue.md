---
title: Durable knowledge and config writes; BOM before frontmatter; escaped index titles
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - storage
  - okf
---

Found in the 0.2.0 codebase audit.

- Knowledge entries, index files and the user configuration were renamed into place without syncing the temp file first (the markdown store syncs), so a crash could leave an empty file.
- An issue file starting with a byte order mark was reported as missing its frontmatter.
- An entry title containing `]` broke its link in the generated index.md.

## Open questions
