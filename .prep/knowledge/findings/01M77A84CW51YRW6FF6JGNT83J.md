---
type: finding
title: 'Derived scope measurement: 42% gain a scope, but explicit scopes already fill the cap'
description: '72 of 171 linked unscoped current decisions gain a derived scope (median 10.5 files, max 17); typical cli/tui files match 16-25 explicit decisions, so the five-decision cap already hides explicit ones: STOP'
date: "2026-11-12"
issue: 01M74VAE370N8SQXGKRVG2B9WM
---

Measured 2026-11-12 on this repository. Current (not superseded) decisions: 312; with an explicit scope other than ** : 119; linked to a done issue with no scope: 171. Derived scope: files of the resolution's evidence commit and of the commit that added resolution.md, without .prep paths, commits over 30 files dropped. 72 of 171 (42%) gain a derived scope, so the one-third condition holds. Size of derived scopes: median 10.5 files, max 17 (min 1; 19 of 72 have 5 or fewer). Typical files, explicit matches (scope globs, directory scopes as prefixes) / derived matches among unscoped decisions: internal/cli/activity.go 16/0, dependencies.go 16/0, handback.go 16/2, hook_shell.go 25/0, setup.go 16/4; internal/tui/actions.go 21/9, app.go 21/30, complete.go 21/0, editing.go 21/8, mouse.go 21/6. Explicit scopes alone are 16 to 25 per file (internal/tui scope 21 decisions, internal/cli 12, internal/cli/hook*.go 9), well over the cap of five, so with explicit decisions ranked first a derived decision would never reach the notice, and the explicit ones are already truncated. Verdict: STOP. The share condition passes; the cap condition fails.
