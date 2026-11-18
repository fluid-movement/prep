---
title: 'Cloud sessions: no session-start activity event, blank request event unexplained'
kind: research
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

In cloud sessions the SessionStart briefing reaches the agent but the activity stream has no `session-start primed` event, and no request or after-tool events appeared in two fresh cloud sessions, although PreToolUse and the prompt hook were recorded. An earlier cloud session showed a `request` event with no value. Find out whether the activity bridge (the plugin's `register.ts`, `prep hook post-model-request`, `prep hook session-start`) really misses events in a cloud session or whether the probes read the stream too early or from a fresh machine-local home, and why one `request` row was blank. Source: the findings "Cloud spike, part 2" and "Cloud probe C/D: hooks and activity" (`prep knowledge find cloud probe`). Answer with a finding; a fix is a separate issue.

## Open questions
