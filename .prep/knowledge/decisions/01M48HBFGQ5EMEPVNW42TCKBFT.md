---
type: decision
title: "Edit frontmatter in place"
description: "Edit frontmatter in place"
date: "2026-10-06"
issue: 01M46PFM9GSM3PYTY4H8M1537D
---

OKF entries may carry keys prep does not model (generated, verified, sources, resource). Updating the YAML node instead of re-rendering from the domain struct keeps them and their order. Alternative: model every OKF key in the domain (grows with the format).
