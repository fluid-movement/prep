---
title: Release 0.4.0 and move the cloud environment pin
kind: manual
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
  - release
priority: high
---

The cloud environment pins the prep release v0.3.0, which lacks changes made since (for example the `session-start primed` activity event, `prep cloud script`, `prep handoff`, the schema message and the docs). Cut a release that contains this work and move the pin:

1. `just release v0.4.0`, push main and the tag, and wait for the release workflow to publish the archives and checksums.
2. Run `prep cloud script` from the new binary and paste its output into the cloud environment's setup script (the dialog at claude.ai/code). This step is the user's.
3. Start one cloud session and confirm `prep version` there and the `primed` event in `prep activity`.

Needs the user for steps 1's push of the tag and for step 2.

## Open questions
