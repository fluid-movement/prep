---
type: decision
title: Cloud handback reports are versioned findings, not work-log text
description: Store the full Markdown report with prep finding under the title Cloud handback <issue-id>; logs flatten Markdown and research handback imports complete issues.
date: "2026-11-18"
issue: 01M7Q2AR66ZVXQKW5H049T44VC
supersedes:
  - 01M7Q6PEKE42FCKQTHC761FN6R
---

Correction to the earlier implementation finding: prep log normalizes all whitespace, so storing a full handback report there destroys its Markdown section structure. The cloud prompt instead asks prep finding <id> --title "Cloud handback <id>" --description "Cloud implementation report" --body-file <report-file>. This preserves the handback format in versioned records without calling prep handback, whose research path would check criteria and complete the issue. Fetch recognizes those linked findings to report presence.
