---
type: decision
title: "Own diff rendering, not a markdown code block"
description: "Own diff rendering, not a markdown code block"
date: "2026-10-06"
issue: 01M46QA9T0FJG0DQ5VNKARFPWE
---

A diff fenced as markdown would be colored by Glamour's syntax highlighter, outside the design tokens. A small LCS line diff rendered by a ui component keeps the colors in the theme. Requirements are short, so an O(n·m) diff is fine.
