---
title: Record human review of knowledge entries
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - knowledge
  - okf
priority: low
---

OKF v0.2 expresses trust with verified events (by, at) next to generated. prep records who generated an entry but not who reviewed it, so agents cannot tell a reviewed entry from an unreviewed draft beyond the status field. When the user approves knowledge changes, typically when an issue completes or a bootstrap draft is reviewed, prep records a verified event with the human actor and the time, and can set a draft to stable at the same moment. prep guide and the TUI knowledge view show the trust tier (generated only, or verified by a person) and can filter by it.

## Open questions

- Is approval its own command (prep knowledge verify <entry>), part of completing an issue, or both?
- Does a later agent update of the entry invalidate earlier verified events, or are they kept as history?
