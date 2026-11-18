---
title: Fail clearly when the repository schema is newer than the binary
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

A cloud session runs a pinned prep binary, and the repository's `.prep` schema can be newer than it. Today nothing says so; the binary may reject the records with a confusing error or, worse, rewrite them in an older shape.

Make a mismatch visible and safe: when a command reads a repository whose schema version is newer than the binary supports, it fails with one clear message naming both versions and the fix (update prep, or move the pin in the cloud environment script). Cover the SessionStart briefing too, so a session starting with an old binary tells the agent and the user at once, instead of showing no briefing.

Decide while defining: whether the repository records a minimum prep version beyond the schema number, and whether `prep check` reports the mismatch.

## Open questions
