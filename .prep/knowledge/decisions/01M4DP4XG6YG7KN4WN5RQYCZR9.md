---
type: decision
title: "Tests: pure units, an SDK session with a scripted model, and a live check"
description: "Tests: pure units, an SDK session with a scripted model, and a live check"
date: "2026-10-08"
issue: 01M46ARM9G3492V298TM82KD45
---

Pure parts (inference, usage areas, view models) run under Node's test runner with type stripping, without a build step. The extension's wiring runs in an SDK session (`createAgentSession`) with pi-ai's faux provider scripting bash calls against a real prep binary in a temporary project: the briefing section, `PREP_ACTOR`, following, refresh after compaction. Drawing is checked by rendering components at fixed widths. Each issue ends with a live check in Pi.
