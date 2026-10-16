---
title: Remove import, artifact, migrate and legacy paths; ack becomes define --trivial
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQKG3PZZ6YS5P4Z81FKGC
  - 01M52ZQGH8HH6DAK9HEGZ28KDE
tags:
  - cli
---

Remove prep import, prep artifact and the artifacts frontmatter, prep migrate and every legacy-record reading path; convert this repository's records once instead. Merge ack into define --trivial (a new baseline that keeps the ready sign-off). Decided in 01M4ZCM9QZ4EQ0STGFX1BKQ2EJ and 01M4ZCMFJ5P67XNSMKX1NPANDF. Delete the code, its tests, fixtures, eval cases, record handling and knowledge text together; nothing stays behind for compatibility. Done when nothing references the removed feature and GitHub CI on main is green.

## Open questions
