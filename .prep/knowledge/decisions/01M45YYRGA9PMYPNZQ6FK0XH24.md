---
type: decision
title: "Golden snapshots include color"
description: "Golden snapshots include color"
date: "2026-10-05"
issue: 01M46QA8TRCHXVESZEA603SJMY
---

Goldens are rendered with a fixed true-color profile, so a changed token shows up as a snapshot diff, not only layout changes. Reviewing a diff of escape codes is noisy, so the gallery is the place to look; the test only guards against unintended change. Alternative: ASCII-only goldens (readable, but blind to color).
