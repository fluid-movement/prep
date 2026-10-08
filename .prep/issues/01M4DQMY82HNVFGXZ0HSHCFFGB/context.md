Claude Code counterpart: `plugins/claude-code/hooks/usage.ts` reads `~/.claude/token-ledger/<session>.jsonl` written by `plugins/token-ledger` (rows per tool call with area, chars, issue; per-request usage). In Pi the extension measures itself:

- Tool output per area: `tool_result` events, classified by `plugins/shared/classify.ts` (Pi tool names), chars of `event.content` text, attributed to the followed issue at that moment (the shared inference).
- Requests: assistant messages carry `usage` (input, output, cacheRead, cacheWrite, cost); read them from `ctx.sessionManager.getBranch()` or the `message_end`/`turn_end` events. Context fill from `ctx.getContextUsage()`, model from `ctx.model`.
- Keep the measurement in memory for the session and rebuild it from the branch at `session_start` (resume); no files written. The per-area numbers must match what token-ledger would record for the same calls (same rules, chars/4 for tokens).
