---
type: decision
title: "Granular commands per record"
description: "Granular commands per record"
date: "2026-10-05"
issue: 01M46P8YDRDX76J95T3VHR41A9
---

Decided with the user: each record gets a command that matches its semantics (decisions appended with generated IDs and dates, criteria added and checked one by one, history appended) rather than one generic write per file. Granular commands let the domain validate meaning, not just format, and keep the API independent of the markdown layout.
