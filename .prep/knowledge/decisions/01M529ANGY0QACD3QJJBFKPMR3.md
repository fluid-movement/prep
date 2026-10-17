---
type: decision
title: "Stay in OKF where possible; document any break"
description: "Stay in OKF where possible; document any break"
date: "2026-10-16"
issue: 01M4ZCMFJ5P67XNSMKX1NPANDF
---

User direction, 2026-10-16. OKF is still beta; staying in the format is preferred, and a break is acceptable if documented and reconciled later. Known tension: OKF requires a generated index.md per directory listing every entry, so new decision records on parallel branches all change the same index and conflict on merge, defeating one-file-per-record. Since the index is derived, first try resolving conflicts by regenerating it (prep fix, or a git merge driver); if that is too clumsy, it is the first candidate for a documented break.
