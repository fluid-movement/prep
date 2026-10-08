# token-ledger

A Claude Code plugin that records where a session's tokens go, so decisions about prep's agent context (what `prep guide` lists, whether a judgment model could pre-filter knowledge entries) rest on measurements.

```
/plugin install token-ledger@prep        # after /plugin marketplace add fluid-movement/prep
claude --plugin-dir plugins/token-ledger # or load it from a checkout
```

It writes one file per session, `~/.claude/token-ledger/<session id>.jsonl`, appended at the end of each turn. Nothing leaves the machine. `/token-ledger` summarizes the current session.

## Rows

Every row has `type` and `t` (milliseconds since the epoch).

| type | fields |
| --- | --- |
| `session` | `cwd`, `isPrep` (a `.prep` directory exists), `interactive` |
| `turn` | `turnId`, `promptChars` |
| `step` | one model request: `turnId`, `index`, `agentId` (subagents), `messages`, `model`, `usage` (`input`, `cacheRead`, `cacheWrite`, `output` tokens), `answerChars`, `tools` (names requested), `stop` |
| `tool` | `turnId` (main loop) or `agentId`, `tool`, `target` (path, command or pattern, clipped to 200 chars), `area` (`prep-cli`, `knowledge`, `issue`, `code`), `prep` (the subcommand), `resultChars` (what the model read back; ≈ chars/4 tokens), `isError`, `ms` |
| `turn.end` | `turnId`, `contextTokens`, `window`, `costUsd` |
| `session.end` | `reason` |

Shell commands are recorded as typed, so a secret on a command line lands in the file; the file stays in your home directory.

Questions the data answers: how many tokens each knowledge read adds and how often it happens per issue, how much of a session's input is prep output versus code, and how context grows turn by turn (`contextTokens`), since every read is paid again on each later request.
