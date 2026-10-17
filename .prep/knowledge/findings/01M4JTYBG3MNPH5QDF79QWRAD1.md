---
type: finding
title: "Findings: Audit the JSON contract against what the TUI shows"
description: "Findings: Audit the JSON contract against what the TUI shows"
date: "2026-10-10"
issue: 01M4JVPQXB93KH02V3WJS1HF0C
---

Everything the desktop prototype needs is reachable; no gap blocks it.

- **Issue list and tree**: `prep list --json` carries id, title, kind, state, stale, blocked, actionable, parent, tags, priority, children count and progress. The tree is built client-side from `parent`. All query flags apply, so filters and saved views (`prep views --json`) work unchanged.
- **Detail**: `prep show <id> --json` carries the requirement, criteria (with checked state), context, context links and paths, decisions, findings, history (work log), baselines, ready, claim, resolution, blocks, children and the effective Definition of Done.
- **Transitions**: `prep guide <id> --json` lists every transition with `allowed`, the exact command, unmet gates (code and message) and the flags it `needs`; the app shows them and runs allowed ones that need no input.
- **Edits**: every write is one CLI command (`edit`, `criterion`, `context`, `decide`, `log`, …); the app runs them with `--root` and `--by`.
- **Knowledge**: `prep knowledge list --json` has entries with sections; `knowledge show <entry> --json` adds `text`.
- **Agent**: `prep activity --json` has events and per-agent focus.
- **Refresh**: `prep watch` prints `{"event":"changed"}` per change, without paths, so the app reloads. A full `prep list --json` takes ~0.1 s on this repository (144 issues) and `show` ~0.1–0.25 s, so a debounced reload of the list and the open issue is fine. Changed paths in watch events would allow incremental refresh; not needed now, left unfiled until a project is large enough to feel it.
