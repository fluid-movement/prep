---
type: component
title: Shared lifecycle hooks
description: 'Shared session, prompt, finish and supporting tool hooks: context levels, record redirects, turn reconciliation, harness protocols and regression tests.'
generated:
  by: codex
  at: 2026-10-13T06:24:37Z
scope:
  - internal/cli/hook*.go
  - internal/userconfig
confirmed_commit: 28f150511800c7e93ff7bcd5e1ed4f499c1f8651
---

# Shared lifecycle hooks

Applies when changing `prep hook` commands, hook context, turn reconciliation, or harness adapters. The implementation is in `internal/cli/hook.go` and `hook_finish.go`, with global choices in `internal/userconfig`.

The hooks serve three tasks: session start identifies the prep project and supplies general current context; prompts steer the agent to search existing issues and use prep tools; finish reviews whether issue progress and knowledge need reconciliation after this turn's changes. Before/after tool hooks support these tasks by redirecting record access and recording activity.

## Commands and protocols

`prep hook session-start|prompt|finish|before-tool|after-tool --harness claude-code|codex` reads one event JSON object on stdin. `--plugin <directory>` supplies the session-start plugin version check. The event's `cwd` identifies the nearest prep project; hook routing ignores `PREP_PROJECT`. Missing projects, malformed/oversized events, integration failures and unavailable Git for turn snapshots fail open. Bad invocation flags still report usage errors; plugin commands suppress errors and handle a missing binary.

- `session-start`: bounded `prep prime` context, at most five items per group and 10,000 UTF-8 bytes. Codex gets an instruction to set `PREP_ACTOR=codex` and the event's `PREP_SESSION` only on prep CLI invocations. Other shell commands keep their usual form; blanket inline environment assignments can prevent approved command prefixes from matching. Hooks do not auto-approve or rewrite ordinary calls.
- `prompt`: default concise guidance on searching before filing, next/guide, ambiguity, reading metadata from the tracker, knowledge maintenance and user scope. Large mode adds guides for at most two distinct full issue ULIDs, at most 4,000 bytes each and 10,000 total. It does not guess suffixes. Small mode does not load the issue/knowledge tree.
- `finish`: compare turn snapshots, then identify knowledge entries whose scopes match changed files. If no issue record changed or matching knowledge entries were not reviewed, request one reconciliation pass, naming up to eight changed files and entries. Claude receives Stop additional context; Codex receives Stop `decision: block` and `reason`. Respect `stop_hook_active` and refresh the baseline before continuing, preventing repeated stop loops. This is a heuristic check, not proof that a record update is semantically complete; the agent decides whether to update, confirm or explain no impact.
- `before-tool`: supported Read/Grep/Glob and shell file reads on prep records return a denial with CLI guidance; Edit/Write/NotebookEdit/apply_patch and common shell writes get write-command guidance. Patch classification uses file headers, not patch body text. Mixed shell calls are checked by segment; a prep call does not exempt a later direct read/write. Normal calls emit no permission decision, preserving harness permissions. Shell parsing covers common literal paths and commands, not every program, variable, symlink, MCP tool or shell construct; this is a guardrail.
- `after-tool`: append actor/session-attributed tool activity, output size and failure state directly from the harness event payload; ordinary command attribution needs no environment assignments. Calls consisting only of prep commands are omitted; mixed code/prep calls remain visible. Claude's module retains environment attribution and request token/context/cost reporting, while tool events come from this command only.

## Global configuration

In `$XDG_CONFIG_HOME/prep/config.yaml` (default `~/.config/prep/config.yaml`):

```yaml
hooks:
  prompt: small
```

`off` returns before event input or project lookup; `small` is the default and injects about one short paragraph; `large` pays for issue loading and guide context. Invalid configuration makes the prompt hook silent. This choice controls prompt context, not session start, record redirects, activity or finish. The hooks never change this file themselves; setup and TUI saves preserve the choice.

## Turn snapshots and limitations

Session start and enabled prompts snapshot Git HEAD and content fingerprints of dirty/untracked files into prep's ignored local `hooks/` state, keyed by harness and a SHA-256 session identifier. Finish compares those fingerprints and commits made during the turn. Unchanged existing dirty work does not trigger review, even if committed during the turn. Files above 8 MiB use size/mtime/mode rather than hashing their entire content; symlinks use their link text. Git calls time out after two seconds. No session ID, no valid snapshot, Git unavailable or snapshot errors means no finish intervention.

With prompt context off, finish uses session-start/the previous finish as its boundary; without a prompt snapshot, it cannot distinguish concurrent outside edits or all continuation changes. Even with prompts enabled, external changes in the same checkout can look like turn changes. Knowledge scopes are hints; reviewing a changed entry or any issue record is detected from content changes, not an LLM judge. Live behavior-suite reruns are required before claiming the original behavior findings are resolved.

## Adapters and verification

Both plugins' `hooks/hooks.json` call these commands directly. Claude also retains the request bridge module. Codex hooks require review/trust after definition changes through `/hooks`; then start a new session. Keep the binary and plugin in sync: old binaries lack `prep hook` and adapters fail open.

Regression tests cover off/small/large, malformed/outside events, context bounds, direct paths, mixed shell calls, patch-body false positives, tool permissions, activity attribution, existing dirty files, committed turn changes, completed reconciliation and stop loops. Run `go test ./...`, `claude plugin test plugins/claude-code` and plugin validation. Actual Claude/Codex sessions and relevant behavior-suite cases remain integration verification, separate from these deterministic checks.

Official protocol references: https://code.claude.com/docs/en/hooks and https://learn.chatgpt.com/docs/hooks.
## Implementation tracking and scope

The user-approved policy is [Tracking implementation work](/decisions/tracking.md): search first, reuse a clear match, create an issue for unmatched implementation of any size unless explicitly opted out, and do not automatically file issues for ordinary questions, exploration or planning. Ambiguous references require clarification before record changes. Planning answers read requirements/metadata and do not infer effort from titles. Stop text explicitly preserves original scope and no-record-write restrictions; automatic continuation grants no extra permission. Shared deterministic checks and targeted Codex fixtures are the current verification gate; original haiku/sonnet thresholds are retained in the deferred Claude verification issue under 0.4.0.
The current small prompt reminder measures 839 UTF-8 bytes and 118 words (2026-10-13). This corrects the 893-byte figure accidentally written in the completed tracking child's final log; it is text size, not tokenizer-based cost.
