---
type: finding
title: Cloud handback 01M7Q68DK55FJW50A8HZBVR2KH
description: Cloud implementation report
date: "2026-11-18"
issue: 01M7Q68DK55FJW50A8HZBVR2KH
---

## Summary
Read `prep activity --json` inside this cloud session (Claude Code 2.1.327, prep plugin 0.3.0, prep binary installed 10:15 UTC) while working. The bridge and hooks do record in the cloud; the missing `session-start primed` is an old installed binary, the blank `request` row is a zero-token request event, and the missing request/summary events are not explained by the code and need a diagnostic follow-up.

## Criteria
1. met — Stream at 12:09 UTC: `session started` (bridge, 12:08:20), `hook prompt sent`, `prep` events (guide, claim, show, knowledge find) and 15 `tool Bash` events (after-tool hook, with `chars`). No `session-start primed`, no `request`, no `summary`, no `gap`, and no `.prep/local/request-gap-*.json` file.
2. met — Reproduced: `printf '{"actor":"claude-code","session":"s","kind":"request","op":"claude-x","tokens":{"input":0,"output":0,"cache_read":0,"cache_write":0}}' | prep activity add` is accepted (rc 0) and `prep activity` shows the row `request  claude-x` with no value. Without `tokens` the event is rejected ("a request event needs tokens, context or cost_usd"). Probe events were removed afterwards from the local stream.
3. met — Causes and proposed fixes are under Findings; the fixes are follow-up issues, no code changed here.

## Findings
### session-start primed is missing only because the installed binary is older than the code
The briefing reached the session, so the SessionStart command hook ran. Piping a SessionStart payload into the installed `/root/.local/bin/prep` (installed 10:15 UTC) printed the briefing but wrote no event. The same payload into `prep` built from this checkout (`go build ./cmd/prep`) wrote `hook session-start primed chars=2829`. `internal/cli/hook.go:135` records it; `git log -S'"primed"'` shows it arrived in d4b5f52 (11:21 UTC), after the installed binary was built, although both report `prep v0.3.0`. The prompt hook records `sent` because it predates that commit. Fix: none in code; the cloud environment's setup script installs a release/older build, so it must install from the checkout (or the version string must change per commit). Follow-up issue for the environment script.

### The blank request row is a request event with all-zero tokens
`plugins/claude-code/hooks/register.ts` (turn.step) emits kind `request` whenever `r.usage` is truthy and fills each counter with `?? 0`. A usage object without counts gives `tokens` all zero; `Event.Validate` (`internal/activity/activity.go:120`) accepts it because `Tokens != nil`, and the activity view prints no value for it. Whether the cloud engine actually returned an empty usage object was not observable (no request event appeared in this session), so the origin of that earlier row stays unconfirmed, but the payload above produces exactly the symptom. Proposed fix: in register.ts, treat usage with no non-zero counts as a `gap` (`op: model`), and/or reject all-zero tokens in `Validate`.

### Request and summary events: none recorded, cause not established
After many model turns there were no `request`, `summary` or `gap` events and no request-gap marker. So `turn.step` never pushed (a gap would be pushed even without usage), `turn.complete` either did not run or failed silently, and `disable()` never fired. Yet `session started` exists, so the module loaded and `session.start` or `turn.start` ran. register.ts hides failures: `turn.complete` and `session.start` use `.catch(($, e, next) => next(e))`, so a throwing `$.session.usage()` would be swallowed with only a debug log. Not provable from inside the session without debug logs. Proposed follow-up: have the bridge write one marker/event on the first `turn.step` and `turn.complete` entry and on any caught error (visible via `prep activity`), then rerun in a cloud session with the right binary.

### after-tool and PreToolUse work in the cloud
`tool` events appear with response sizes, so the PostToolUse command hook runs, and the prompt hook recorded `sent`. Earlier probes' claim of no after-tool events is not reproduced here; the activity stream is `.prep/local/activity.jsonl` inside the container (gitignored), so a probe that reads it from a different container or too early sees an empty or partial stream.

## Problems
- Request/summary absence is unresolved; the follow-up needs bridge diagnostics.
- Probe events written during the investigation were removed from the local stream.
