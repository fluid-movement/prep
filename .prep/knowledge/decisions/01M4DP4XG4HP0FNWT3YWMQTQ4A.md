---
type: decision
title: "Distribution: git source pinned to the release tag until 1.0"
description: "Distribution: git source pinned to the release tag until 1.0"
date: "2026-10-08"
issue: 01M46ARM9G3492V298TM82KD45
---

`prep setup` installs `git:github.com/fluid-movement/prep@vX.Y.Z`; a root `package.json` (private, `pi-package`, a `pi` manifest) points at `plugins/pi`. This mirrors the Claude Code marketplace pin and needs no new infrastructure; Pi clones the whole repository per install, which is accepted for now. At 1.0 the package moves to npm (lighter installs, the pi.dev gallery), decided by the user.
