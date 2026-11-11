---
title: Hand off by brief in the rules, skill and guide, after a trial of brief and handback on real issues
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
depends_on:
  - 01M72CWZ12ESB9D853ENN2NQ67
tags:
  - handoff
priority: high
---

The workflow rules, the skill and the guide's hand-off text tell an orchestrator to delegate with `prep brief <id> --out` and a one-line prompt that points the implementer at the brief file, and to finish with `prep handback <id>`. The implementer and checker passes stay separate. `prep guide` remains the way for an agent that works alone. The rule "pass these rules on to every subagent" no longer applies to implementers and checkers.

Before the guidance changes, `prep brief` and `prep handback` are tried on at least three real issues with a small model as implementer and as checker. The reports it writes are kept as test fixtures for the handback parser. If the reports do not keep to the format, or the agent does not write the report to the named file, the format or the brief's wording changes first and the trial is repeated. The trial needs the user's OK, as live eval runs do.

Why: decision 01M72CVCP45FN7D515257JH6VE (hand-off by brief) and its addition 01M74V5WKJV2D9E2T6J6945EZR, which separates the commands, the trial and the guidance change so that guidance follows evidence. In the daypack sessions the orchestrator wrote 39 subagent prompts and prompt files for 68 `codex exec` runs by hand, each restating the prep protocol (finding 01M72A9TBJ5JQXT2VGT15N2HW1).

## Open questions
