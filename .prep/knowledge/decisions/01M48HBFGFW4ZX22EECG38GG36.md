---
type: decision
title: "Default commit mode stays off"
description: "Default commit mode stays off"
date: "2026-10-06"
issue: 01M46ARS5RNN4X2TMVHAKWA3KW
outcome: true
---

prep stages its writes and the user or agent commits them together with the code, so one commit carries a change and its records, history stays readable, and squash merges need nothing special. all remains available per project. Alternative considered: all by default (exact audit trail in git, at the cost of many tiny commits; the records already carry by and at).
