---
type: decision
title: "Loader injected, no storage imports in the TUI"
description: "Loader injected, no storage imports in the TUI"
date: "2026-10-06"
issue: 01M46ARFD8BQNT8QBJYRB30C35
---

The TUI gets a load function and a directory to watch from the CLI, so it uses exactly the CLI's loading path and stays independent of the markdown and OKF adapters. Alternative: the TUI opens the stores itself (duplicates opening logic, couples it to the adapters).
