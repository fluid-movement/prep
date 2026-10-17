---
type: decision
title: Keep setup.Unregister as test support
description: Unregister isolates tests from the user's real harness registrations; it stays though no command calls it
date: "2026-10-17"
issue: 01M52ZS90G529Q80FGM51RR5DS
---

The prune survey listed setup.Unregister as used only by its own tests. Its callers are test isolation (setup_test register cleanup, cli TestSetup unregistering claude-code and codex), not tests of it; removing it would mean rewriting that isolation for no gain. deadcode -test reports nothing unreachable.
