---
type: decision
title: "Filters intersect with the tab"
description: "Filters intersect with the tab"
date: "2026-10-06"
issue: 01M46QA9T0FJG0DQ5VNKARFPWE
---

A filter narrows what the tab shows. Concatenating flags into one ParseFilter call would OR repeated flags (--state in the tab and in the filter) and widen the result, so the two queries run separately and intersect. Alternative: replace the tab query with the filter (loses the view the user chose).
