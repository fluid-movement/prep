---
type: finding
title: Newer-schema sessions report the missing briefing without writing
description: Session-start hooks surface E_SCHEMA as a single-line additionalContext message naming the project and supported schema numbers, while still exiting 0. The shared schema error directs local users to prep update and cloud users to update the PREP_VERSION pin in the environment setup script, printable with prep cloud script. Commands rejected with E_SCHEMA do not append failure activity, preserving the incompatible project. prep check retains its existing schema diagnostic.
date: "2026-11-18"
issue: 01M7Q2AQYM5W9XZXK9BJ43SZYM
scope:
  - internal/cli/**
---


