---
title: Decision entry format in decisions.md
kind: decision
---

Settle the format of decisions.md entries, including the outcome marker for decision issues. The first implementation uses '## <id>: <title>' headings followed by date:, optional supersedes: and outcome: true lines, then rationale and alternatives. Confirm or change it.

## Open questions

- The parser requires date: on the line right after the heading; a blank line in between (common markdown style) is reported as "decision D1 has no date", which reads as if the date were missing. Allow the blank line, or keep the strict form with a clearer message?
