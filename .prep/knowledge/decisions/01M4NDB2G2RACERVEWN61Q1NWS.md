---
type: decision
title: "A permanent project ID is the identity; the remote is only a clone hint"
description: "A permanent project ID is the identity; the remote is only a clone hint"
date: "2026-10-11"
issue: 01M4N5ARVE6J5APZYJ7VA0N464
supersedes:
  - 01M4NDB2G13X6N4HRV5DFWAK4Y
---

The remote changes when a repository is renamed, transferred or moved to another host, and nothing in the referring project can notice, so it cannot be the identity. Each project gets a ULID `id:` in project.md, written once by `prep init` and committed; links store the target's ID plus its last known remote as a clone hint, under the referring project's own local name. Records use `<local name>:<full issue ID>`. Renaming the target (`name:`), changing its remote or moving its directory never breaks a link; renaming a local name goes through `prep link rename`, which rewrites every reference. A fork that keeps project.md is a clone of the same project; `prep init --new-id` splits it off. The user asked for links "as solid as possible".
