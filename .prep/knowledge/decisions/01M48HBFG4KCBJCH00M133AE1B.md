---
type: decision
title: "Write injected like Load"
description: "Write injected like Load"
date: "2026-10-06"
issue: 01M46ARGCGTMC4GMFJYRWB8098
---

The TUI receives a write function that runs the CLI's pipeline (plan, CheckWrite, Apply, stage), mirroring the injected loader. It keeps the TUI free of storage imports and guarantees TUI writes pass the same validation as CLI writes. Alternative: the TUI shells out to the prep binary (slower, stringly-typed errors, harder to test).
