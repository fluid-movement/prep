The 162 issue decisions and the findings.md files are converted mechanically by the knowledge stage of 01M52ZRB3VY0MSZ0HZWJ2SEC24, in every prep project; this issue converts only this repository's 43 descriptive entries, moves the overview to AGENTS.md and rewrites principle 9.

Approved conversion tables, filled in batch by batch as the user reviews them. Apply only the rows listed here; every other statement of the entries in a reviewed batch is deleted. All 43 entries are reviewed (batches 1 to 4, 121 records). From batch 3 on, the user delegated the review to the main session (2026-10-17): it decides, records the result here, and brings only real questions to the user.

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

### Approved batch 2 (2026-10-17): components markdown-store through tui-editing
Approved: the rows below, with scope paths corrected; dropped markdown-store 1, okf-store 1 and 2, tui-editing 2 (covered by batch 1 or removed with drift and confirm) and tui-design-system 2 (merged into palette 1). Amend batch 1 'Write pipeline holds one lock': each file is written through a temporary file and a rename. tui-agent and okf-store convert to nothing.

## /components/markdown-store.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 2 | Constraint | State-defining file written last, removals after | Apply writes supporting records first, then baseline, ready.md and claim.md, then resolution.md, and removes files after every write. | (implied) an interruption leaves the previous state, or the new state with a leftover file, never a state without its supporting files; complete writes resolution.md before removing claim.md. | internal/mdstore |
| 3 | Constraint | Skip temp-prefix files on load | Names starting with `.prep-tmp-` are ignored in .prep and issue dirs. | One may belong to a write in flight or be left by a crash. | internal/mdstore |
| 4 | Decision | Config: own file or committed dist, never merged | The user's gitignored config.yaml wins as a whole over committed config.yaml.dist; no merging. | reason not stated | internal/mdstore, internal/domain |
| 5 | Constraint | Store validates themes, domain does not | Theme names are checked in the store (load P003, write refused), not in PlanConfig. | Domain cannot import palette (import cycle). (Same source as palette.md; merged.) | internal/mdstore, internal/palette |
Deleted: Load/format walk-through, BOM tolerance, requirementBody, record-write mechanics, priority rendering, ViewOrder, legacy 0.1.0 config conversion (no compatibility before release), renderConfig details, DefaultConfig content, Init file list, Fmt/Fix lists, contract corpus note, project identity parsing, completion-record paragraph.

## /components/okf-store.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
Deleted: Walk/frontmatter parsing, Entry.Body, link rules, git NUL/optional-locks detail (= batch 1), index file generation, frontmatter key order, reserved names.

## /components/palette.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Decision | Eight palette tokens only, none for single uses | Palette has text, muted, subtle, accent, selection, success, warning, error; borders, states and kinds derive from them; no token is added for a single use. | Keeps the UI calm; a new state or kind needs no palette change (01M4AM1CEQQPWSNCP05Y6SVD7T); merged from tui-design-system. | internal/palette, internal/tui/theme |
| 2 | Decision | Themes are made dark first | A theme is one palette for dark terminals; only `default` and `light` are built in. | Nearly every developer runs dark (01M4AMC70GBW1R8YDA1TG44HKE). | internal/palette |
| 3 | Decision | Themes set no background | Selection colors are a step lighter than common dark backgrounds and the theme never sets a background. | Selection must show on whatever background the terminal uses. | internal/palette |
Deleted: Color struct and downsampling, Resolve/Validate/Builtins API, fallback by terminal background, dependency note (merged into markdown-store record 5), derived color mapping (state colors rationale "color marks what needs attention, finished work fades" is kept as Doubtful), screen path.

## /components/projects.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Decision | Project identity is the ULID, not name or remote | A project is its permanent ID in project.md; name is a handle and the remote only says where to clone. | Renames, remote changes and moves keep identity; a fork stays the same project until `prep init --new-id` (01M4N5ARVE6J5APZYJ7VA0N464 D2). | internal/projects, internal/mdstore |
| 2 | Decision | Registry is machine-local state outside user config | projects.json lives in XDG state, not in the user configuration. | It is machine-local state, not a choice. Damaged file is treated as empty and rebuilt. | internal/projects, internal/userconfig |
| 3 | Constraint | Recording a project never fails or slows a command | Registry touch happens after success, at most once a minute per root, and failures are ignored. | (implied) the registry is a convenience; a burst of commands costs one write; lost touches self-repair. | internal/projects, internal/cli |
Deleted: Rows, Describe, listing/JSON shape, Resolve order, commands, TUI client, tests env note, stat-error handling.

## /components/release.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Constraint | Archive names are a contract | `prep_<version>_<os>_<arch>.tar.gz` (zip on Windows) plus checksums.txt must match update.ArchiveName and install.sh. | Both consumers derive the name. | .goreleaser.yaml, internal/update, install.sh |
| 2 | Constraint | Releases are tag-driven and versions must agree | Pushing a v* tag releases; the workflow first checks every plugin/marketplace manifest carries the tag's version. | (implied) marketplace pinned to the release tag (= batch 1 in part). | .github/workflows, justfile |
| 3 | Decision | install.sh never edits shell files and needs no root | Installs into ~/.local/bin and prints a ready-to-paste PATH line instead of editing rc files. | XDG user executable location, no root needed; (implied) editing user shell files is intrusive. Also: downloads are checksum-verified, and update refuses on mismatch. | install.sh |
Deleted: GoReleaser matrix, just release steps, version stamping rules, IsRelease details, update mechanics, Windows rename, verification notes, setup call after install.

## /components/setup.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Decision | User config stores choices only | userconfig records choices (hooks.prompt, harnesses, tui.*); installed versions are always asked from the harness. | Store only what cannot be derived (design principles; "a small database"). | internal/userconfig, internal/setup |
| 2 | Constraint | One failing harness never blocks the others | Setup apply/refresh handles harnesses independently and reports a result each. | (implied) integrations are independent adapters. | internal/setup |
Deleted: Harness interface, setup flags, tty reading under curl|sh, callers, integration pointers, no AGENTS.md needed.

## /components/tui-agent.md
Deleted: key map, data source, labeling, view lists, usage/feed/liveliness description. "Harnesses draw no panels" is = batch 1 (no harness-specific panels). Default view Issue ("the current issue matters more than the tool calls") is Doubtful.

## /components/tui-design-system.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Constraint | Screens never set colors or styles | Screens compose ui components; colors come from theme tokens, never literals at a use site. | Calm, consistent look; change colors only in internal/palette/builtin.go. | internal/tui |
| 3 | Decision | Spacing tokens, no screen margin | Use SpaceS/M/L and Row, not literals; no margin around the whole screen. | Cells are not pixels; a margin costs four columns and two rows at 80x24 (01M4EDFAWDFFX96Y4ZQCKS9BQ3). | internal/tui/ui |
| 4 | Constraint | Screens work at 80x24 | Every screen and component must fit 80x24. | (implied) minimum supported terminal. | internal/tui |
| 5 | Constraint | Every component variant is in the gallery | A new ui component takes the theme, adds all variants to Gallery, and updates goldens; check both themes. | Visual changes are reviewed through snapshots. | internal/tui/ui |
Deleted: theme resolution and rebuild mechanics, components list, layout helpers, gallery/snapshot test descriptions.

## /components/tui-detail.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Decision | Children shown as one heading with a list-all link | Detail shows a single Children heading with progress and `list all`, a few unresolved children, resolved ones only in the progress. | Handles any number of children (30+) without flooding the detail (01M4G1BS7D78QHNNQE9963B259). | internal/tui |
Deleted: relation block layout, link mode keys, tests list, children view mechanics. "Tab never moves between links" is Doubtful.

## /components/tui-editing.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | Decision | Keys are melodies: unshifted letter sequences | Every TUI action is a short sequence of unshifted letters, never a chord, except ctrl+s/ctrl+e in text fields. | Decided in 01M47Y6EA0SMCZAKG03785YF76; reason not stated in the entry (memory: stable mnemonics, arrows only as fallback). | internal/tui |
| 3 | Decision | Action menu lists only what no key of its own does | Each action has exactly one sequence; the menu shows only actions that apply now. | So each action has one sequence (01M4EDFAR0SWRETYBKJ5WC3TRP). | internal/tui |
Deleted: tag dialog, key tables, keymap/help, dialog mechanics, create wizard, editor hand-off, settings pane. Completion rules by kind and human actor are covered by today's changes.

### Approved batch 3 (2026-10-17, decided by the main session): TUI components, watch, features
Dropped: watch 1 (= tui 1), tui 4 and cross-project 3 (behavior description), cli-contracts 1 (= the 'MCP parked' decision of 01M4ZCM7TTPASZ2EN8MQN6EPVZ, converted with the issue decisions). tui-mouse 3 includes the user's rule that the wheel scrolls a list's view, never its selection. agent-context, completion-policy, knowledge-lifecycle and knowledge-policy convert to nothing.
## /components/tui-filter.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Live filter bar parses list flags | The TUI filter bar takes `prep list` flags plus bare words (joined into one text phrase), filters live on every parseable edit, and a word starting with `-` is always a flag, never a title search; flag values come from `Tree.QueryFlags`, the same source as `prep flags`. | So live filtering does not jump on the first `-`; one value source keeps TUI and CLI in step. | internal/tui, internal/domain |
| 2 | constraint | Filter only narrows | A tab's filter is intersected with the tab's own query, so a filter can never widen a view; esc restores the filter from when the bar opened. | reason not stated | internal/tui |
| 3 | decision | `/` is an unlisted alias of `f` | `/` also opens the filter bar but stays out of the keymap. | `/` is a chord on German layouts, kept for US-keyboard habits. | internal/tui |

Deleted: completion picker mechanics (candidates, ordering, counts, which flags are offered), footer flag names, per-tab/title display, error-line behavior, pointer to mouse doc.

## /components/tui-knowledge.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | TUI knowledge screen is read-only | The TUI only shows knowledge; nothing in it edits entries. | Agents write knowledge for agents. | internal/tui |

Deleted: layout, keys, meta line, "Changed by"/`o` links, attention marks and Options.Check (K003-K006 drift/confirm = knowledge drift, DELETE), issue-detail "knows" links.

## /components/tui-mouse.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | constraint | Mouse mirrors keys | Keys stay the primary way in; every click does exactly what the clicked element's key does, and a click on the selected item does what enter does. | reason not stated | internal/tui |
| 2 | decision | Mouse capture is the user's choice | Capture is `tui.mouse` in user config (unset = on), not a project setting; with it off the terminal selects text with a plain drag and the keymap/footer never list mouse actions. | (implied) the terminal setup is per person, not per project. | internal/tui, internal/userconfig |
| 3 | decision | Click selects, click again acts; the wheel scrolls the view | Selecting and activating are separate clicks with no timed double click, and the wheel scrolls a list's view, never its selection. | User decision (TUI mouse rules). | internal/tui |

Deleted: marking/IDs/hit-testing layers, back button handling, per-element click lists, wheel offsets and `wheeled` mechanics, tests.

## /components/tui-navigation.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Two-tier navigation, one grammar | A screens bar (project name plus Issues, Agent, Knowledge) sits over the current screen's views; both tiers use plain labels, current one accent and bold, others muted, with no underline, chips, separators or status dots; the project name is plain text because it is not a place. | Decided in 01M4EA4RB500QF2449Y3914VMQ and 01M4EDFAWDFFX96Y4ZQCKS9BQ3; the accent marks places only. | internal/tui |
| 2 | decision | Back stack restores screen and selection | Backspace (or the mouse back button) returns to the previous screen with its selection, from anywhere; there is no breadcrumb in the bar. | reason not stated | internal/tui |
| 3 | decision | Short terminals merge the tiers | Below 16 rows both tiers share one line and wide-and-short keeps the side-by-side split. | So the TUI fits in a strip above or below the agent. | internal/tui |

Deleted: project switcher mechanics (Switch/Previous restart), key list (= approved "keys are unshifted letter sequences"), click ids, spacing (= approved "spacing tokens"), tests, `headerH`.

## /components/tui.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Live reload splits tree from activity | A project change reloads the issue tree; a `.prep/local` change only re-reads the activity stream; a load error keeps the last good tree and shows the error; selections are kept by issue ID. | So a busy agent never makes the TUI reparse every issue. | internal/tui, internal/watch |
| 2 | decision | Viewports scroll vertically only | Left and right keep their navigation meaning, so text viewports never scroll horizontally. | Stated: left/right keep navigation meaning. | internal/tui |
| 3 | constraint | Fresh stores per load | The CLI builds fresh stores for each TUI load and check because they can run concurrently. | Stated: reloads and checks can run concurrently. | internal/cli, internal/tui |

= approved: entry/no storage-adapter imports (package import boundaries); `tui.layout` single/split as user config (user config stores choices only); dark-first startup and retheme; short ID = ULID suffix.
Deleted: view/tab/list/detail rendering, tags, priority, stale diff, screen pointers, layout numbers, key inventory, tests, Bubble Tea notes, check screen drift loading.

## /components/watch.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 2 | decision | Watcher errors mean full reload | An overflow or failed watch is reported as a change of everything. | Reload everything rather than miss changes. | internal/watch |
| 3 | decision | Debounced, merged, never blocking | Events wait 150 ms because a transition writes several files; one goroutine owns pending state and a new report merges into the waiting one. | So a slow reader never blocks the watcher. | internal/watch |

Deleted: fsnotify recursion handling, tests.

## /features/agent-activity.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Activity is local UI state | The stream lives in `.prep/local/activity.jsonl`, per machine and per worktree, git-ignored, never a record, never staged or committed. | It is UI state; reads still never change records. | internal/activity |
| 2 | constraint | Appends are single small writes | Each event is one JSON line written with `O_APPEND` in one write under 4 KiB (strings clipped, oversize refused), and the file is rewritten to its newest lines past 1 MiB. | So concurrent agents never interleave lines. | internal/activity |
| 3 | decision | Record writes always, reads only for a named actor | Prep commands log writes always but reads only when run with `--by` or `PREP_ACTOR`. | Panels, hooks and people reading without a name buried the agent's own events. | internal/activity |

Also kept inside rec 3 as one sentence: focus follows prep events naming an issue except `show`, and never events with `project` or harness tool events (so a write in another project leaves the home focus alone). Reason stated.
= issue decision: ended or 30-minute-quiet agents hidden, linked session IDs count as one agent (and keep the predecessor's focus); filters never alter stored events.
= approved: Codex after-tool reporter, PREP_SESSION/PREP_ACTOR, hooks fail open, `.prep` edit guard, no rewriting/auto-approval of calls.
Deleted: event field list, session event validation, kinds/readers, Claude and Codex bridge descriptions, missing Codex token fields, P007 detail.

## /features/agent-context.md
Deleted: guide behavior (covered by today's completion changes and DoD guidance), "agents execute project checks" (kept once in cli-contracts), knowledge update on completion (drift/review gate).

## /features/cli-contracts.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 2 | decision | `prep guide` is the lifecycle contract | Failed transitions point back to `prep guide <id>`, and domain planners decide validity and unmet conditions. | One shared source of next steps. (implied) | internal/cli, internal/domain |
| 3 | constraint | Prep never runs project checks | A report of a passing check is an agent assertion; prep only verifies what records show. (Merged from completion-policy, lifecycle, agent-context.) | reason not stated | internal/domain, internal/cli |

= approved: JSON response fields and exit codes (0/1/2/3).
Deleted: artifact add/remove and frontmatter (removed), `confirmed_commit` (review marker), documentation-decision completion, additive-field compatibility text.

## /features/completion-policy.md
Deleted: all of it is covered by today's changes (DoD guidance, any actor, decision/research from defined, parents from ready), the removed docs gate, or kept once elsewhere (no project checks in cli-contracts, commit-format-only in lifecycle). `prep dod` editing is behavior.

## /features/cross-project.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | `--project` targets any registered project | A global flag (or `PREP_PROJECT`) takes name, ID suffix, remote or path; every command works there, writes stage in the target and never commit, output names the target; flags beat env vars, both root flags are `E_USAGE`, ambiguous or unknown projects fail before the command runs. | reason not stated | internal/cli, internal/projects |
| 2 | decision | Activity copy to home stream | A command in another project logs to the target's stream and a copy (with `Project` set) to the home stream; copies never move focus. | A write elsewhere must leave the home project's current issue alone. | internal/activity, internal/cli |

Deleted: TUI project switching (in tui-navigation, not a decision), skill rule 11 text, resolveProject mechanics. (Project links and cross-project dependencies are removed, not touched here.)

## /features/knowledge-lifecycle.md
Deleted: documentation decision gate (`--docs`/`--no-impact`), review against scoped code, DoD pointer.

## /features/knowledge-policy.md
Deleted: superseded by decision and finding records; the briefing-first rule is = approved "small global prompt context"; DoD as guidance covered today.

## /features/lifecycle.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | State is derived, never stored | Issue state comes from which files exist (resolution, claim, ready against newest baseline, baseline); no state field is written. | reason not stated | internal/domain, internal/mdstore |
| 2 | decision | Define settles what, enrich settles how | The user has authority over the requirement at define; at enrich the agent drafts context, decisions and criteria, the user agrees, and `prep ready` signs off; code may be read during define only to test whether the requirement makes sense. | reason not stated | internal/domain |
| 3 | decision | Completion evidence is format-checked only | Without `--commit` the evidence is the commit that adds `resolution.md`, so code and records land together; a given hash is validated for format only. | Evidence stays valid after a squash merge removes the branch commit (01M4AVPH9A56WJ4JFMYZT476H0, 01M46ARQ78P8M40K4FSPVNK531). | internal/domain, internal/cli |
| 4 | decision | Prep never pushes; claims travel with branches | A claim is visible elsewhere once its branch merges; working on main has no lag. | Decided in 01M46ARR6G4K7NZBYZQJZGBCR0. | internal/domain |
| 5 | constraint | Record writes and edits never change state | `context`, `decide`, `criterion`, `findings`, `log` and `edit` work on unresolved issues without moving state; only a requirement or kind edit can make a defined issue stale. | reason not stated | internal/domain, internal/cli |

Deleted: transition table and gates (code; completion rules covered by today's changes), ack (now `--trivial`, covered today), DoD, actor gate, stale/actionable/blocked definitions, reopen/drop/release details, I023, parents never claimed (covered).

## /features/local-write-safety.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | constraint | No unlocked fallback | The advisory lock is `.prep/local/write.lock`; a lock access failure is surfaced, never bypassed, and reads outside a prep project never create `.prep`. | Process death releases the lock; a silent fallback would lose safety. (implied) | internal/mdstore |
| 2 | decision | TUI writes reload under the lock | The TUI builds a fresh store per write, holds the lock, loads again, then plans and applies. | A long-running TUI must never plan on a stale screen. | internal/tui, internal/cli, internal/mdstore |
| 3 | decision | No journal, no rollback | An interrupted write leaves the previous or new state plus harmless extra files that `prep check` reports; the fix is to run the command again. | (implied) ordered atomic file writes make recovery by rerun enough. | internal/mdstore |
| 4 | constraint | Bounded guarantee | Safety covers cooperating prep processes on one checkout, not direct editors, power loss or distributed coordination; other worktrees have separate records, and parallel agents stay the orchestrator's job. Markdown stays authoritative, domain policy stays out of storage. | Stated limits. | internal/mdstore |

= approved: one lock per write command (no stale-plan hash), temp-plus-rename, state file last and removals after (complete writes resolution then removes claim), skip `.prep-tmp-` files.
Deleted: Hold/Locked mechanics, writeRank detail, tests.

### Approved batch 4 (2026-10-17, decided by the main session): conventions, decisions, overview
Dropped: knowledge-base 2 (records are named by ULID) and 4 (= principle 10); storage-format 1, 2, 4 (= batches 1 and 2), 5 (= principle 10), 6 (issue decisions become knowledge records). Principle 8 'no index or cache' is superseded by the record converted from storage decision D1 of 01M4ZCMFJ5P67XNSMKX1NPANDF (a derived, disposable SQLite index); principles 1 and 3 and the Beads lesson stand. The overview converts to AGENTS.md (draft below) and no records.
All records are kind decision. "Why" paraphrases the entry; "(implied)" marks inference.

## /conventions/knowledge-base.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Knowledge bundle follows OKF | The knowledge bundle follows OKF closely: prep generates index.md (never edited), the root declares okf_version, log.md is reserved. | So OKF tools can read the bundle. | internal/okf |
| 3 | decision | One concept per entry, kept small | One concept per entry, small; lint warns above 8 KiB. | (implied) cheap targeted retrieval | internal/okf |

Deleted: bootstrapping, confirm and drift, completion-time documentation review, branch and backlink notes, "current state only", precedence rule, code-inventory guidance, draft/stable status, review experiments, the old type list. All depend on removed features or describe behavior.

## /conventions/storage-format.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 3 | decision | Per-machine state lives in .prep/local | Machine-specific state (activity stream, write lock) goes in a git-ignored local directory. | (implied) it must not be versioned or shared between clones | internal/mdstore, internal/activity |
| 7 | decision | Canonical file form enforced by prep fmt | Files have one canonical form (LF, fixed key order, one field per line, single blank lines); prep fmt normalizes and CI runs fmt --check. | (implied) clean diffs | internal/mdstore |
| 8 | decision | A heading appears once per record file | A "##" heading appears at most once per record file; prep check reports repeats. | Stated: prep reads only the first copy of Open questions. | internal/mdstore |

Deleted: exact frontmatter field lists and allowed-file lists (describe the code; check enforces them), error code details, tag and priority syntax, baseline file naming, ack and resolution documentation fields, artifacts, confirmed_commit, write-lock and tmp-file details (covered by local-write-safety), Definition of Done location (now guidance).

## /decisions/design-principles.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Enforce steps, not how work happens | prep defines what each step must produce and the gates between steps, not how the work is done. | reason not stated | |
| 2 | decision | Write permissions belong to the harness | prep has no approval mechanism; reads and writes are separate commands so harness rules can tell them apart. | Stated: harness rules must distinguish reads from writes. | internal/cli |
| 3 | decision | Store only what cannot be derived | Store every fact that cannot be derived and nothing that can; state, progress, blocked and actionable are computed. | "It is a small database." | internal/domain |
| 4 | decision | Every transition leaves a record in its file | Each transition is recorded in the file that represents it; no fact depends on git history. | reason not stated | internal/domain, internal/mdstore |
| 5 | decision | Files never move | An issue lives at its ID directory forever. | reason not stated | internal/mdstore |
| 6 | decision | Storage and presentation are decoupled | Files are storage, git is versioning, CLI and TUI are query interfaces. | reason not stated | internal/cli, internal/tui |
| 7 | decision | Domain and storage are separate layers | Domain logic is independent of storage; markdown is the first storage adapter. | reason not stated | internal/domain, internal/mdstore |
| 8 | decision | Keep it simple: scan all files, no index or cache | Scan all files on every run; add no index or cache until proven necessary. | "Keep it simple." | internal/mdstore | superseded by storage D1 of 01M4ZCMFJ5P67XNSMKX1NPANDF
| 9 | decision | Issues are the work; knowledge is what was decided and learned | Issues track the work; the knowledge base holds only decisions and findings. | Rewritten per today's decision, replacing "issues record change, the knowledge base records state". | |
| 10 | decision | The CLI is the API | Every data change goes through a command carrying content, not files; hand edits stay possible and prep check catches them. | Stated: agents and clients never need hand edits. | internal/cli |

Deleted: preamble about the original design document; link from principle 9 to the old conventions entry.

## /decisions/out-of-scope.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | No GitHub integration | prep has no GitHub integration. | reason not stated | |
| 2 | decision | Single project only | prep covers one project; there is no cross-project view. | reason not stated | |
| 3 | decision | prep never orchestrates implementation | prep records work but never runs agents. | reason not stated | |
| 4 | decision | No approval or permission system | Harness permissions handle approval. | See principle 2. | |
| 5 | decision | No focus, goal or ranking between issues | No active goal or goal ordering; priority is fixed levels per issue, no ranking between issues (since 01M4874AR04JFPZ5KYW0402JC1). | reason not stated | internal/domain |
| 6 | decision | How work inside a step happens is not prescribed | It is the user's and agent's choice. | See principle 1. | |
| 7 | decision | User config holds only personal choices | ~/.config/prep holds only user choices such as integrated harnesses; project behavior stays in project config. | reason not stated | internal/userconfig |
| 8 | decision | Hosted or database storage deferred | Possible later through the storage port; markdown only for now. | reason not stated | internal/mdstore |
| 9 | decision | Editing, visualizing and publishing knowledge is left to other tools | Left to OKF ecosystem tools and $EDITOR; human-facing docs could be generated from knowledge later. | reason not stated | |

Deleted: "completion evidence is a plain commit reference" clause (completion contract removed); link to the Harness setup entry.

## /decisions/prior-art.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Taken from and avoided in Backlog.md | Took binary-served instructions, a project-level Definition of Done and dogfooding; avoided titles in file names, status as a field, scanning branches for cross-branch state. | reason not stated | |
| 2 | decision | Taken from and avoided in Beads | Took the session-start briefing (prime), JSON output everywhere and contract tests; avoided syncing two representations (SQLite and JSONL) and many records in one file. | reason not stated | |

Deleted: opening summary sentence (names drift detection, removed), spec-driven tools remark (not a choice).

## /decisions/stack.md
| # | Kind | Title | One-line description | Why (from the entry) | Scopes |
|---|---|---|---|---|---|
| 1 | decision | Go, one static binary | Core, CLI and TUI ship as one static Go binary. | Near-instant startup for frequent agent calls, simple install; TypeScript needs a runtime and starts slower, Rust has no decisive advantage and slower iteration. | go.mod, cmd |
| 2 | decision | TUI on Charm v2 | The TUI uses Charm v2 libraries under charm.land at latest releases; v1 import paths are not used; pin majors when adding a Charm library. | Moved from v1 in 01M48N7E9GRPWH1SKX5MTB931Y; otherwise reason not stated. | internal/tui, go.mod |
| 3 | decision | Third-party storage adapters speak JSON-RPC over stdio | External adapters use JSON-RPC over stdio, like LSP and MCP. | So the core stays a closed, tested binary. | internal/domain |
| 4 | decision | Harness adapters are thin TypeScript only where required | TypeScript only where a harness requires it (Claude Code mod, Pi extension). | (implied) the harness requires it | plugins |

Deleted: dependency list, binary-size and startup figures (drift).

## /overview.md
Content goes to AGENTS.md. Deleted: entry index (generated index replaces it), "completion requires a documentation decision", original-design-document note.

AGENTS.md draft (repository root; replaces /overview.md):

````markdown
# AGENTS.md

prep is a workflow engine and memory for coding agents, with a human supervising. It is a Go CLI and TUI that keeps issues, their lifecycle records and a knowledge base of decisions and findings in the repository under `.prep/`, versioned with the code. This repository is prep itself.

## Where things live

- `cmd/prep`: the binary's main package.
- `internal/domain`: the core model, lifecycle rules and ports; no storage knowledge.
- `internal/mdstore`: the markdown storage adapter for `.prep/`.
- `internal/okf`: the OKF knowledge bundle (entries, index generation).
- `internal/cli`: commands and output; `internal/cli/skill.md` is the source of the prep skill.
- `internal/tui`: the terminal UI; `internal/palette`: its colors and themes.
- `internal/activity`: per-machine agent activity stream; `internal/watch`: file watching.
- `internal/gitx`: git access; `internal/projects`: project registry.
- `internal/setup`: harness integration install (`prep setup`); `internal/update`: self-update; `internal/userconfig`: `~/.config/prep`.
- `internal/eval`, `eval/`: evaluation of prep with agents.
- `plugins/`: the Claude Code and Codex plugins; `testdata/`: the contract corpus and fixtures.
- `install.sh`, `justfile`, `README.md`, `TOKEN-EFFICIENCY.md`.

## Build, test, check

- `go test ./...` (or `just test`) runs all tests, including the contract corpus.
- `just install` builds with the version stamped, installs the binary and refreshes the harness integrations from this checkout.
- `just lint` runs gofmt and go vet; `just fmt` formats Go and the prep records.
- `prep check` validates the records; `prep fmt --check` verifies canonical form. `just check` runs both from source; `just ci` runs lint, test and check.
- After changing `internal/cli/skill.md`, run `just sync-skill` so the plugin copies match.

## Work and decisions

- Work on this repository is tracked with prep itself. Start an issue with `prep guide <id>`; `prep prime` gives the project briefing.
- Change prep records only through prep commands, never by editing their files.
- Decisions and findings live in the knowledge base. Find them with `prep knowledge find <words>` and read them with `prep knowledge show <entry>`; record new decisions there, not in code comments or docs files.
- Commit straight to `main` unless told otherwise.
````
