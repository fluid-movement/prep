---
type: decision
title: "A resolved issue's knowledge references are history"
description: "A resolved issue's knowledge references are history"
date: "2026-10-17"
issue: 01M52ZQ7J4SGWXY2W9X8EPEJV1
---

User agreed, 2026-10-17. Deleting the desktop entry made prep check report I020 errors in 8 done issues whose resolutions named it, and I025 warnings in their contexts. A resolution records what was true when the issue closed; rewriting it to satisfy a check would rewrite history, and the knowledge conversion will delete many entries. I020 is removed (it only ever checked resolutions); I025 skips resolved issues; prep complete still refuses --docs entries that do not exist. Alternative rejected: editing old resolutions.
