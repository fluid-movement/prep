---
type: decision
title: "Edits append a history line"
description: "Edits append a history line"
date: "2026-10-05"
issue: 01M46P8XEGAVJ634ZN0E50MSV9
---

Every write leaves a record without relying on git history (design principle 4). The line names the time, the actor and the changed fields, not old and new values; requirement changes are visible through baselines and staleness. Alternative: no record (loses who changed relations and titles).
