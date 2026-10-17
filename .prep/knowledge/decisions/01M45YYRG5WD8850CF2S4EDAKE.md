---
type: decision
title: "--depends-on replaces the list"
description: "--depends-on replaces the list"
date: "2026-10-05"
issue: 01M46P8XEGAVJ634ZN0E50MSV9
---

One flag with replace semantics, decided with the user: the given values become the whole list, a single empty value clears it, and an empty value mixed with IDs is a usage error. Alternative: separate add and remove flags (two flags for one field).
