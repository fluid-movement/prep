---
type: decision
title: Tracking implementation work
description: 'User-approved tracking policy: search before implementation, reuse clear matches, file unmatched work, honor opt-outs and preserve scope through finish reviews.'
generated:
  by: cli/prep-v0.2.0-72-gf273360
  at: 2026-10-13T09:44:17Z
scope:
  - internal/cli/hook.go
  - internal/cli/skill.md
  - internal/cli/similar.go
  - eval/cases/natural
confirmed_commit: 22a58e58c0cf17a7b64b7477344c30797dc116cf
---

# Tracking implementation work

Applies when changing shared hook guidance, the prep skill, or natural-task eval expectations. The user agreed this policy in the untracked-work decision issue on 2026-10-13.

In an initialized prep project, search for existing issues before implementation. Reuse a clear match and follow its current guide, lifecycle, knowledge and acceptance gates. If no issue covers authorized implementation, create one first regardless of change size. A one-line behavior fix is still implementation; guessing whether work is substantial is not a reliable tracking rule.

Honor explicit tracking opt-outs and no-record-write restrictions. Ordinary questions, exploration and planning do not automatically require issues. When an issue is named, inspect it and its current guide rather than file another. When multiple issues plausibly match a reference, ask which one before changing any of them.

The original request defines scope throughout the turn. Updating a requirement does not authorize implementing or completing that issue. An automatic finish review grants no new permission and never overrides earlier restrictions; report needed follow-up when a record update would exceed scope.

`prep new` surfaces up to five similar unresolved issues based on title-word overlap with existing titles and requirements. Hints are advisory: inspect the candidates, do not automatically choose or merge one, and do not treat their absence as proof that no issue exists. Creation remains possible for intentionally distinct work. Search before calling new; a hint after creation cannot prevent every duplicate.

Shared session and prompt guidance implements the policy for every adapter. Finish reviews source changes against knowledge scopes and issue progress, with a single scope-preserving continuation when needed. See [Shared lifecycle hooks](/components/hooks.md) and [Agent context](/features/agent-context.md).

The natural no-issue eval now expects one new issue for unmatched implementation. Separate explicit-opt-out and read-only cases expect no new issues, and ambiguous-reference cases require leaving candidates unchanged until clarified. Original haiku/sonnet thresholds are preserved in the deferred Claude verification issue under 0.4.0. Current shared-mechanism regressions and targeted Codex checks do not establish success rates for other models or harnesses.
