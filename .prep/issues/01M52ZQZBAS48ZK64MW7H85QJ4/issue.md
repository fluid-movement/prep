---
title: 'Record layout: no empty files, union merge for the work log'
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQSEP2EKTYKYFC79EGFS6
tags:
  - storage
---

prep new and later commands create a record file only when something is written to it; remove the empty files in this repository. Add merge=union for history.md in .gitattributes so parallel appends on branches do not conflict, and have prep init write it. Decided in 01M4ZCMFJ5P67XNSMKX1NPANDF.

## Open questions
