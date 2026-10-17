---
type: decision
title: "Sort at the query layer, not in Tree.IDs"
description: "Sort at the query layer, not in Tree.IDs"
date: "2026-10-06"
issue: 01M4874AR04JFPZ5KYW0402JC1
---

Tree.IDs stays chronological (validation, indexing and stable diagnostics rely on it). A domain helper orders IDs by priority then ID, applied where work is listed: Tree.Query results, the roots and children in Tree.TreeOrder, and prep prime's lists. Alternative: make IDs priority-ordered (reorders unrelated output such as diagnostics and changes the children index).
