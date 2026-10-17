---
type: decision
title: "Subprocess with test-only env vars, not in-process"
description: "Subprocess with test-only env vars, not in-process"
date: "2026-10-12"
issue: 01M4QKVB5EW2N6Y9BDVCFJDC7Q
---

The builder runs the real binary so setup is exactly what an agent's calls would be and the activity stream carries the setup actor. A process cannot hold a clock between calls, so prep reads `PREP_TEST_NOW` and `PREP_TEST_SEED` and the builder advances the time itself. Alternative considered: in-process calls with the existing package-level `clock`/`entropy`; rejected because it skips the binary's own wiring and cannot be reused by `prep eval` against a pinned binary.
