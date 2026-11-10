---
title: guide carries the issue's own records at every step, so no show follows it
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
priority: high
---

`prep guide <id>` prints the issue's requirement (with its open questions), acceptance criteria and context whenever they exist, at every step and for every kind, state and parent. Today it does so only for code issues that are ready or in progress (the implementer's brief); in every other case it points at `prep show` instead.

Why: agents follow that pointer. In the daypack and fpa-events sessions of November 2026, `prep show` followed a guide without the requirement 48 of 59 times, and a guide with it 3 of 34 times (finding 01M72A9TBJ5JQXT2VGT15N2HW1). Each of those shows is an extra command, and often an extra request that re-reads the whole context.

The pointers to `issue.md`, `acceptance.md` and `context.md` leave the guide's Read section when their content is printed. The Definition of Done and the hand-off instructions stay where they are now. `prep show` is unchanged.

## Open questions
