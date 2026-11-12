---
type: finding
title: 'Real-world run: scoped-decision case does not fail without the before-tool hook'
description: 'Real claude -p sessions (haiku, real plugin hooks incl. UserPromptSubmit), scenario eval/cases/knowledge-read/scoped-decision-at-edit.yaml in fresh scratch repos /tmp/liveA/{with,without}{1,2,3} (streams /tmp/liveA/*.jsonl; runner /tmp/liveA/run.sh). With before-tool hook: 3/3 wrote Discounted with the 500 floor (hook notice reached the model at the first Edit in 2 of 3; the 3rd looked the decision up itself first). Without it (prep shim answering nothing for hook before-tool): also 3/3 satisfied the rule, each ran prep knowledge find/show itself within its first 2 tool calls, prompted by the UserPromptSubmit nudge (search prep before implementing). Criterion 8 is not met in the real plugin setup: the case only fails without the hook when the prompt hook is also absent (the eval skill-only variant); the hook was not what made the difference here. Caveat: N=3 per arm, haiku.'
date: "2026-11-12"
issue: 01M72ACGWREA9E1TFNGTJB0A98
---


