---
type: decision
title: "Render through the adapter before validating"
description: "Render through the adapter before validating"
date: "2026-10-06"
issue: 01M46PFM9GSM3PYTY4H8M1537D
---

Links are extracted by the OKF parser, so the adapter renders the would-be entry and parses it back; the domain validates the tree with that entry before anything is written. The domain keeps the rules, the adapter keeps the format. Alternative: move link parsing into the domain (couples it to markdown).
