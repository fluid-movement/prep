---
type: decision
title: "Keep a --no-bootstrap escape hatch"
description: "Keep a --no-bootstrap escape hatch"
date: "2026-10-06"
issue: 01M47YPB38JZ3KPVNG0CWB146M
---

prep init creates the bootstrap issues by default, as the requirement says. A --no-bootstrap flag skips them for projects that do not want a knowledge base yet and for tests; prime still reports the missing overview and prep knowledge bootstrap creates the issues later.
