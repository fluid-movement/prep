---
type: decision
title: "Themes are validated in the store, not in PlanConfig"
description: "Themes are validated in the store, not in PlanConfig"
date: "2026-10-07"
issue: 01M46QWSKGQZ5S0N3RHTEE86Y1
---

`palette` imports `domain` (for the state and kind token names), so `domain.PlanConfig` cannot call it. The markdown store validates themes when it loads a config (P003) and refuses to write an invalid one (E_USAGE); `domain.ThemeDef` is plain data.

Rationale: keeps domain free of UI vocabulary without a registration hook, and every write already goes through the store.

Alternatives: a validator function registered into domain at init (implicit wiring); hard-coding states and kinds in palette so domain can import it (duplication).
