Approved conversion tables, filled in batch by batch as the user reviews them. Apply only the rows listed here; every other statement of the entries in a reviewed batch is deleted. Batches not yet listed are still to be drafted and reviewed.

### Approved batch 1 (2026-10-17): components architecture through hooks

Approved by the user: the drafted rows below, with architecture 4 and cli 1, 6, 7 and 8 dropped (covered by design principles 2, 4 and 6, the tracking and completion decisions, or removed with prep import). 'Implied' reasons were inferred by the drafter and accepted as such. Everything not listed is deleted.

## /components/architecture.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | constraint | Domain owns lifecycle rules | CLI and TUI share one domain model and transition planners; storage adapters load and persist but never implement lifecycle rules. | So both interfaces agree on derived states and transitions (stated). | internal/domain/**, internal/mdstore/**, internal/okf/** |
| 2 | decision | Separate issue and knowledge stores | IssueStore (checked record writes) and KnowledgeStore (retrieval, Render then Apply) are separate ports. | Access patterns differ (stated). | internal/domain/ports.go |
| 3 | constraint | Writes are planned, then checked, then applied | Every write is a domain.Change; Tree.CheckWrite validates it in memory and rejects writes that add errors before the adapter renders and writes. | Reason not stated. | internal/domain/**, internal/cli/** |
| 5 | constraint | Keep package import boundaries | activity and palette import no store/TUI, eval does not import cli; leaf packages stay leaves. | Reason not stated. | internal/activity/**, internal/palette/**, internal/eval/** |
| 6 | decision | Harnesses are adapters, workflow logic is shared | Lifecycle hook behavior lives once in internal/cli/hook*.go; Claude Code and Codex provide only event adapters. (also hooks.md) | (implied) one behavior for all harnesses | internal/cli/hook*.go, plugins/** |

Deleted: package/file map, port method listing, "CLI and TUI share Filter and Query" (follows from #1), cross-reference lines.

## /components/claude-code-panel.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | No harness-specific panels | The Claude Code side panel was removed; prep tui --agent runs next to any agent and prep records agent actions itself. | Rebuilding the TUI in each harness was the wrong layer (stated). Dated: decided in 01M4DTQZPMFCKN0DHDWA95VKTN. | plugins/claude-code/** |

Deleted: description of the removed panel and its commands.

## /components/claude-code.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | constraint | One skill source, copies kept identical | internal/cli/skill.md is the source; the Claude plugin skill and .claude/skills/prep/SKILL.md must stay identical (just sync-skill, tested). Codex copy is generated from it plus Codex notes. (merges cli.md, codex.md) | Reason not stated. | internal/cli/skill.md, plugins/*/skills/**, .claude/skills/** |
| 2 | decision | Bridge sets PREP_SESSION and PREP_ACTOR | The Claude bridge sets session and actor env at session start and on session ID change so Bash prep commands and bridge events count as one agent. | So reads are recorded without --by and commands and events meet as one agent (stated). | plugins/claude-code/hooks/** |
| 3 | decision | Tool activity is reported only by after-tool | prep hook after-tool is the single reporter of tool events; the bridge reports only token usage, context fill and cost; prep-only calls are omitted because prep records them itself. (merges codex.md, hooks.md) | prep records its own commands (stated); single source avoids double counting (implied). | plugins/**/hooks/**, internal/cli/hook*.go |
| 4 | decision | Marketplace pinned to the binary's release tag | Install adds the marketplace at the binary's tag; both plugin manifests carry the same version and the release refuses a tag that differs. | Third-party marketplaces do not auto-update (stated). | internal/setup/**, .claude-plugin/**, plugins/** |
| 5 | constraint | Plugin refresh must keep sibling plugins | Moving the pin uninstalls every plugin of the marketplace, so refresh reinstalls the others, restores the old marketplace on failure and warns. (merges codex.md) | Removing the marketplace uninstalls all its plugins (stated). | internal/setup/claudecode/**, internal/setup/codex/** |
| 6 | decision | Cloud sessions use a static project skill | Claude Code on the web loads no user plugins, so the repo keeps .claude/skills/prep/SKILL.md as a project skill. | Cloud sessions load no user plugins (stated). | .claude/skills/** |

Deleted: plugin file listing, hook behavior description, version-drift warning (see Doubtful), shared-workflow summary, tests, install command details, cloud research issue pointers.

## /components/cli.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 2 | constraint | JSON output contract and exit codes | Under --json all output, errors included, is JSON on stdout; exit codes are 0 ok, 1 gate/check failure, 2 usage, 3 conflict. | Reason not stated (agents parse it). | internal/cli/** |
| 3 | decision | Arguments after -- are literal | From -- on, arguments are positionals, so text may be "--json". | So free text can look like a flag (stated). | internal/cli/** |
| 4 | decision | prep stages but never commits | Writes stage their paths except git-ignored ones (user's own config); the user or agent commits .prep with the code. | Ignored config must never fail git add (stated); not committing reason not stated. | internal/cli/write.go, internal/gitx/** |
| 5 | constraint | Write pipeline holds one lock | lock, load, Plan, CheckWrite, Apply, stage: the checkout lock covers load through the last write. | (implied) no lost updates | internal/cli/write*.go |
| 9 | decision | Priority and tags sit outside the baselined requirement | Priority (unset = medium) and tags never make an issue stale and may be edited on resolved issues. (also domain.md) | (implied) they are metadata, not requirement | internal/domain/** |

Deleted: flags list, edit/record/knowledge/setup/update/tui/skill command listings, bootstrap and init mechanics, --text matching, show JSON fields, test seams, reopen, human: actor rule (see Doubtful).

## /components/codex.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| - | = merged | skill generated from skill.md | claude-code.md #1 | | |
| - | = merged | PreToolUse guardrail and attribution env only on prep CLI | hooks.md #2 and #3 | | |
| - | = merged | after-tool activity, prep calls omitted | claude-code.md #3 | | |
| - | = merged | fail open, hook errors do not block | hooks.md #1 | | |
| - | = merged | refresh keeps sibling plugins | claude-code.md #5 | | |

Deleted: package layout, context injection description, install steps, trust notes, live-verification advice, attribution regression list, legacy launcher (see Doubtful), "no AGENTS.md needed" (see Doubtful).

## /components/domain.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Issue IDs are ULIDs, resolved by unique suffix | IDs are ULIDs (same-millisecond IDs increment to keep order) and any unique suffix, either case, resolves. | String order equals creation order (stated); ID format decided in 01M485SBR0WN8JDQSN7A3GT7CB, reason there. | internal/domain/ulid.go, internal/domain/tree.go |
| 2 | constraint | Diagnostic codes are stable and never reused | P*, I*, K* codes with severity and class are permanent. | Reason not stated. | internal/domain/validate.go, internal/domain/diag.go |
| 3 | decision | Filter semantics: AND across flags, OR within, not- excludes | Different flags AND, repeated flags OR, list values may exclude with not- or !; tags may not start with not-; state values use hyphens. | Reason not stated. Decided in 01M4ANA516QWSG4BAQE0QCCR9F. | internal/domain/query.go |
| - | = merged | Priority/tags outside baseline | cli.md #9 | | |

Deleted: file-by-file tour of the package (types, tree, fsm, records, import, bootstrap, knowledge, config, flags, knowview, knowread, guide, errors), test mechanics, bootstrap refusal (see Doubtful), resolved-issue reference handling (see Doubtful).

## /components/eval-cases.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Eval cases come from real sessions and must discriminate | Cases are designed from real sessions, and cases every model passes trivially are dropped. | They cost money and show nothing (stated). | eval/cases/** |
| 2 | decision | Split plugin variants to measure each part | Variants skill-only and hook-only, against none and current, show what skill and hooks each buy. | Answers what each part buys (stated). | eval/variants/** |

Deleted: case family catalogue, errors command anecdote, tracking-policy expectations (already /decisions/tracking.md), "no live Claude runs authorized" (status).

## /components/eval.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Deterministic checks first, judge last | Assertions are mechanical wherever possible; an agent judge is used only for what cannot be checked mechanically. | (implied) cost and reliability | internal/eval/**, eval/** |
| 2 | decision | Fixtures are built by the real prep binary | Fixture states are reached via real transitions (--by fixture), not hand-written files. | (implied) fixtures stay valid as rules change | internal/eval/fixture/** |
| 3 | constraint | Every run is isolated | Each run gets its own HOME, Claude config, registry and PATH with a pinned binary copy; Leaks compares init against the variant. | (implied) no contamination of runs or the user's setup | internal/eval/env/** |
| 4 | constraint | Harness failures are never failed assertions | Errors (including is_error with subtype success) get error status and are excluded from pass rates, shown as an errored count. | (implied) do not blame the agent for harness failure | internal/eval/run/**, internal/eval/matrix/** |
| 5 | decision | Judge rules and calibration | One call per rubric item, pinned model, no tools, errors never count as pass; results are untrusted when hand-label agreement is below a threshold. | Reason not stated. | internal/eval/judge/** |
| 6 | decision | Transcripts are harness-neutral | Claude stream-json is parsed to a neutral transcript so another harness needs only an adapter. | Stated. | internal/eval/transcript/** |
| 7 | decision | Differences count only when intervals do not overlap | Pass rates use Wilson intervals; Compare reports a difference only when they do not overlap. | (implied) runs are noisy | internal/eval/results/** |
| 8 | decision | Evals run locally only | Evals run via just eval-smoke/eval-full, never in CI. | (implied) credentials, cost | justfile, internal/eval/** |

Deleted: layout and command list, credentials how-to, per-package descriptions, models command, not-covered list, retired-contract note, workspace cleanup details.

## /components/git.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | constraint | Git calls never take the index lock | Every git call sets GIT_OPTIONAL_LOCKS=0. | So reads like status from the TUI or prep check never block a commit the person runs at the same time (stated). | internal/gitx/** |
| 2 | constraint | Path lists are NUL-separated | Read paths with -z, never the line form. | The line form quotes non-ASCII and special characters that never match real paths (stated). | internal/gitx/** |
| - | = merged | Stage only tracked paths | cli.md #4 | | |

Deleted: function inventory, usage list, test note.

## /components/hooks.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | constraint | Hooks fail open | Missing project, malformed events, missing binary, git or snapshot errors, and a failed bridge send let the call through; only deliberate direct-record denials block. (merges claude-code.md, codex.md) | (implied) a prep fault must not stop the agent | internal/cli/hook*.go, plugins/** |
| 2 | decision | Direct .prep access is redirected to the CLI | before-tool denies supported reads and writes of prep records with CLI guidance; it is a guardrail, not enforcement. | (implied) records must go through checked writes | internal/cli/hook*.go |
| 3 | constraint | Hooks never touch ordinary permissions | Normal calls get no permission decision, are never rewritten or auto-approved; Codex env assignments only on prep CLI invocations. | Blanket inline env assignments stop approved command prefixes from matching (stated). | internal/cli/hook*.go, plugins/codex/** |
| 4 | decision | Prompt context is global, small by default | Config hooks.prompt off/small/large; small is default; large adds guides at a cost; invalid config makes it silent; hooks never edit the file. | Large pays for issue loading and guide context (stated); small keeps cost down. | internal/cli/hook.go, internal/userconfig/** |
| 5 | decision | Finish asks for one reconciliation by heuristic | Stop compares turn snapshots and scoped knowledge, then asks once for issue/knowledge reconciliation; stop_hook_active prevents loops; no LLM judge, agent decides. | Heuristic by design (stated); reason for no judge not stated. | internal/cli/hook_finish.go |
| 6 | constraint | Continuation grants no extra permission | Stop text preserves original scope and no-record-write restrictions; automatic continuation adds no permission. | Reason not stated. | internal/cli/hook_finish.go |
| - | = merged | Shared hook logic | architecture.md #6 | | |
| - | = merged | after-tool single reporter | claude-code.md #3 | | |

Deleted: command descriptions, byte bounds, snapshot mechanics and limits (8 MiB, 2 s), adapters and test list, protocol links, tracking policy text (/decisions/tracking.md), prompt byte count.
