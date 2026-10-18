---
type: decision
title: Activity stream archives instead of pruning
description: Past its size limit the stream is archived to a dated file for later usage analysis; readers see only the current stream
date: "2026-10-18"
issue: 01M571MHR3K5Q695DTJCZ9XCSP
supersedes:
  - 01M529ANGRZSGRVBGGKV09XNES
---

Decided with the user for the usage analysis after 0.3.0, 2026-10-18. The stream stays the Agent screen's live feed, and readers still see only the current stream. Past 1 MiB, though, it is archived to `.prep/local/activity-<date>.jsonl` and a new stream starts, instead of being pruned to its newest lines, so that weeks of use can be analysed afterwards. prep never reads or prunes the archives. Hooks record their outcomes as `hook` events and failed prep commands record their error code, for the same analysis.
