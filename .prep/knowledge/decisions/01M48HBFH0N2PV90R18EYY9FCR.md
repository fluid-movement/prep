---
type: decision
title: "Config writes go through the domain like issue writes"
description: "Config writes go through the domain like issue writes"
date: "2026-10-06"
issue: 01M47Y6EA0SMCZAKG03785YF76
---

Settings edits are a domain Change (PlanConfig) validated by CheckWrite and written by the markdown store, so the TUI uses the same pipeline (stage or commit via record) as every other write, and an invalid view query is rejected before anything is written. Alternative: the TUI writing config.yaml directly (bypasses validation and the commit mode).
