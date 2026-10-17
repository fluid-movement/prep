Where things are:
- Event kinds and validation: `internal/activity/activity.go` (`Kind*` constants, `Event`, `Validate`, which today accepts only tool and request from harnesses). Add a session kind with started/ended and an optional previous session ID.
- Agent list and focus: `Agents` and `Foci` in `internal/activity/activity.go`. Both resolve a chain of linked session IDs to one agent; `Agents` leaves out sessions with an ended event and sessions whose newest event is older than 30 minutes (a fixed constant; take the current time as a parameter so tests are deterministic).
- Callers: `internal/tui/agent.go` (the Agent view's list, footer count 'agent n/m' and tab cycling), `internal/cli/activity.go` (`prep activity add`), `internal/cli/read.go` ('Recently worked on' in the session briefing).
- Claude Code bridge: `plugins/claude-code/hooks/register.ts`: `session.start` sends started, `session.end` sends ended (before the existing queue flush), and `turn.start` on an ID change sends ended for the old ID and started with previous for the new one. Keep the bridge's rule that sends are queued, never awaited, and failures let the call through.
- Tests: Go tests for validation, `Agents` (ended, quiet, linked chain) and `Foci` across a linked chain; the TUI agent view's tests and goldens if the count changes (regenerate with the tests' update flag and review the diff); the bridge's tests next to `register.ts` if they exist.
- The activity file is local (`.prep/local/activity.jsonl`, not committed). Do not edit it; the checker trims this repository's stale `live-demo-1` and actor-keyed lines.
- Knowledge: [Agent activity](/features/agent-activity.md), [TUI Agent screen](/components/tui-agent.md), [Claude Code integration](/components/claude-code.md). Update through `prep knowledge update`.

Checks before stopping: `go test ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`, and the bridge's tests if it has any.
