---
outcome: done
by: cli/prep-v0.3.0-3-gb07f9b4
at: 2026-11-11T08:33:21Z
evidence: d4b5f52
---

Tool and hook events carry the Claude Code agent id; prep-command events do not, because the binary sees nothing in its environment that names the subagent (the report attributes them by session). Codex tool hooks carry no agent id. Archives are read with --all, not --since. A project without git records no snapshot error.
