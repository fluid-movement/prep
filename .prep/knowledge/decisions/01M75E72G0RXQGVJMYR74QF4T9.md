---
type: decision
title: Before-tool hook names scoped decisions at the first edit
description: Edit tools get a note naming up to five current decisions scoped to the file, once per session and agent; the Stop question no longer lists them.
date: "2026-11-11"
issue: 01M72ACGWREA9E1TFNGTJB0A98
scope:
  - internal/cli/hook.go
---

Agents skipped the lookup rule (knowledge lookup before 0-29% of code-changing turns); prep knows the scoped decisions, so it names them at the first edit instead of after the change in the Stop question. The note carries title, ID and description cut at 150 characters so it is usually enough without a prep knowledge show. Most specific scope first, then newest; at most five, with a count of the rest. Scope ** is left out (the session briefing lists it). Shown IDs are kept per harness, session and agent_id under .prep/local/hooks and cleared at session start, so compaction shows them again and subagents get their own list. The hook adds context only and never blocks; it fails open. Outcome is a before-tool hook event with verb context. Codex gets the same output but delivery to the model is unconfirmed, so it is documented as Claude Code only. Edits through shell commands are out of scope.
