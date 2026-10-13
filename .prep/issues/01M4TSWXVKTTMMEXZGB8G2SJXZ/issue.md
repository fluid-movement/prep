---
title: Unknown cross-project dependencies and explicit review prerequisites
kind: code
parent: 01M4TST3C4ZSZDKA5WKBA3ND4H
tags:
  - projects
---

Represent cross-project dependency status as satisfied, unsatisfied, or unknown, with source checkout/revision and uncertainty visible to the user. Unknown dependencies block claiming by default; an explicit human-approved override must record a reason. Implementation completion satisfies dependencies by default; allow a dependency to require human approval instead. Do not infer satisfaction merely because a project is not checked out locally. Reuse existing project identity, links, and dependency machinery from 01M4N5ARVE6J5APZYJ7VA0N464; remote synchronization is not required. Prioritize after the completion/review and TUI slice.

## Open questions
