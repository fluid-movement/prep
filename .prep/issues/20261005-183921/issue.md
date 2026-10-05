---
title: prep check flags duplicate section headings
kind: code
---

A record file with a repeated section heading, such as two "## Open questions" headings in issue.md, passes prep check today; the import of the design doc left one behind unnoticed. prep check reports a duplicate section heading in any record file as a diagnostic with a stable code, and prep fix merges the duplicate when the extra section is empty. The contract corpus gets a broken tree for the case.

## Open questions

## Open questions
