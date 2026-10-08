# Token efficiency: measuring before integrating Jev

Working notes for one experiment: find out where an agent's tokens go when it works through prep, then decide whether integrating Jev, or something simpler, would cut them. Nothing here is decided yet; the data decides.

## What we are trying to achieve

Agents working on prep issues should spend fewer LLM tokens, with no loss in quality. Speed does not matter here; tokens do.

The question is whether [Jev](https://www.eesel.ai/blog/typesafe-jev) (TypeSafe's judgment model) can take work off the LLM. If it can't, the question becomes which other changes to prep would save tokens.

## What Jev can and cannot do

- Jev answers typed judgments only:
  - **yes/no**: the probability that a statement is true;
  - **choice**: one option from a list you supply (up to 255);
  - **score**: a position on a scale you define.
- Each answer carries a confidence. Below a threshold the question goes back to the LLM. The [`jev-use`](https://pi.dev/packages/jev-use?name=jev) npm package wraps this for Claude Code, as an MCP tool, a CLI and a library.
- Jev cannot write text or code. TypeSafe calls it unreliable for counting, arithmetic, date ordering and numeric precision; compute those in code.
- Cost: about $0.04 per million input tokens, output free; latency about 230 ms p50. These are third-party figures, so check before relying on them.
- Jev is in early access. Whatever it judges is sent to TypeSafe or to a gateway (OpenRouter, Vercel).
- **It is a rate win, not a token win.** jev-use's own docs say a judgment costs *more* tokens than the LLM deciding in place. LLM tokens are saved only when the data being judged never enters the conversation. In prep, that means the binary or a script calls Jev and the agent sees only the verdicts.

Most of an agent's work in prep is writing (requirements, context, criteria, code, documentation), and Jev cannot do any of it. Jev can only help with the reading and triage around the writing.

## Where we start from

Measured on this repository on 2026-10-08:

| What the agent reads | Size |
| --- | --- |
| `prep prime` | ~430 bytes (~110 tokens) |
| `prep guide <id>` | ~1.2 KB (~300 tokens) |
| One knowledge entry | 1–8 KB (~300–2,000 tokens) |
| The whole knowledge base (21 entries) | 75 KB (~19k tokens) |

prep's own output is already small. The candidate token sinks are:

1. knowledge entries the agent opens;
2. source code the agent reads while implementing;
3. how both pile up in the context. Every read is paid again, as cache reads, on every later request in the session.

## How we measure

Two plugins in this repository record and show the data:

- **token-ledger** (`plugins/token-ledger`) writes `~/.claude/token-ledger/<session id>.jsonl`. It records:
  - every model request, with input, cache read, cache write and output tokens;
  - on each request and tool call, the prep issue being worked on, so sessions that cover several issues split by issue;
  - every tool call, with what it touched, its prep area (`knowledge`, `issue`, `prep-cli`, `code`), the prep subcommand, and how much text it returned;
  - context fill and cost at the end of each turn.

  The row format is in its README. To load it from a checkout: `claude --plugin-dir plugins/token-ledger`.
- **The prep pane's Usage tab** (`/prep:pane usage`) shows the current session's figures as they accumulate.

Record ordinary prep work, unchanged, for one to two weeks. Aim for at least 10 issues taken through enrichment or implementation.

## What we compute

| Metric | Definition |
| --- | --- |
| **K-share** | Knowledge-entry tokens read back ÷ all tool-output tokens, per prep session |
| **K-per-issue** | Knowledge tokens read while working on one issue (rows' `issue`) |
| **Tokens per issue** | Model tokens per issue, to find which issues, kinds or phases cost the most |
| **K-unused** | Share of knowledge reads that did not feed the work (proxy below) |
| **Amplification** | For each read: its tokens × the model requests that followed it in the session; the read's real cost, mostly as cache reads |
| **Code-share** | Code tokens read back ÷ all tool-output tokens |
| **prep-share** | prep command output ÷ all tool-output tokens, by subcommand |
| **Drift cost** | Tokens spent re-reading entries and diffs for `prep check` drift warnings (K005) and `prep knowledge confirm` |

**K-unused is a proxy.** A read counts as unused when all of these hold:

- the entry is not linked from the issue's `context.md` or its parents' context;
- its `scope` does not match the files the session edited;
- the session never updates the entry.

Spot-check about 10 sessions by hand to see whether the proxy matches what really happened.

## What an outcome would mean

| Outcome | What it means | What we do |
| --- | --- | --- |
| K-share < 10% and K-per-issue < 3k | Knowledge reads are cheap; Jev has nothing worth filtering | **No Jev.** Leave knowledge retrieval alone. |
| K-share ≥ 25% and K-unused ≥ ⅓ | Agents open entries they don't need, and that costs real tokens | **Prototype the Jev relevance filter** (idea J1) as a script, then compare the same issues with and without it. Try the Go-only trims (G1–G3) alongside; if they get most of the gain, prefer them. |
| K-share ≥ 25% but K-unused low | Knowledge is read and used; it is doing its job | **No Jev.** Make reads cheaper instead: section pointers (G2), smaller entries (G3). |
| Code-share dominates (≥ 60%) and K-share is small | Agents rebuild understanding from code that the knowledge base should hold | **No Jev.** The opposite fix: better entries, or more of them, scoped to the code agents keep re-reading (G5). Spending more on knowledge reads to save on code reads is the trade. |
| prep-share ≥ 10%, or one subcommand stands out | prep's own output is too verbose | **No Jev.** Trim that command's output (G4). |
| Drift cost ≥ 5k tokens per week | Drift checks are a real sink | **Prototype Jev drift triage** (idea J2). |
| Amplification is large for early reads | Early reads are paid for many times over | Prefer reading late, and push wide reads into subagents, which keep them out of the main context (G6). Applies with or without Jev. |
| None of the Jev outcomes hold | — | Close the Jev question, record it in `.prep/knowledge/decisions/`, and keep only the Go-side improvements the data supports. |

These thresholds are starting points, not commitments. Adjust them once the first numbers show what typical sessions look like.

## Ideas that use Jev

Ranked by expected LLM-token savings. All of them follow the same rule: the data goes from prep to Jev, and only the verdicts reach the agent.

- **J1. Relevance filter for `prep guide`.**
  - What: send each candidate entry's full body and the issue's requirement to Jev, and get a relevance score per entry. `guide` lists the relevant entries first and collapses the rest.
  - Saves: knowledge reads the agent doesn't need, times their amplification. Likely the largest saving, since every implementation starts with `guide`.
- **J2. Drift triage (`prep knowledge triage --drifted`).**
  - What: for each drifted entry, send the entry and `git diff confirmed_commit..HEAD -- <scope>`, and ask whether the diff contradicts anything in the entry. A confident no gets `prep knowledge confirm`; anything else goes to the agent.
  - Saves: most drift re-reads, when drift is frequent.
- **J3. Documentation decision at `prep complete`.**
  - What: given the diff and the candidate entries, Jev picks no impact, update entry X, or new entry, and prep prefills the suggestion.
  - Saves: re-reading candidate entries to decide.
- **J4. Small decisions:** stale ack-or-define, kind/priority/tags/parent on `prep new`, duplicate check in `prep import`, advisory requirement and criteria checks.
  - Saves: little. The LLM usually has the facts in context already, so these save a few hundred tokens at most. Only J4's import duplicate check moves bulk data out of the context.

Constraints that apply to every Jev integration:

- Advisory only. Jev verdicts never decide a transition; state and gates stay deterministic (design principles 1 and 3).
- An optional port in the domain with a provider adapter. It is off by default and enabled per project, because repository content leaves the machine.
- Prototype as a script first, piping `prep … --json` to `jev-use judge`. Build into the binary only what the A/B comparison proves.

## Token savings that need no Jev

- **G1. Trim the candidate list in `prep guide`.** Stop listing the overview on every issue, show scope matches first, and cap the list.
- **G2. Point to sections, not entries.** With stable section headings in entries, `guide` can point to `entry.md#section`, so the agent reads only the relevant part.
- **G3. Smaller entries.** Lower the lint warning from 8 KiB to about 4 KiB (`components/tui.md` is 8.2 KB today) and split entries that cover more than one concept.
- **G4. Trim verbose command output** wherever prep-share shows a command is expensive. Pointers over content, as `prime` and `guide` already do.
- **G5. Knowledge that replaces code reading.** Where the data shows agents repeatedly reading the same code to understand it, write or extend a scoped entry. One cheap read can replace several expensive ones.
- **G6. Read late, read narrow.** Skill guidance: open knowledge only when the step needs it, prefer `Grep` and offset reads to whole files, and send wide surveys to subagents so their reads never enter the main context.
- **G7. Read knowledge through prep.** `prep knowledge list|find|show` (01M4DA304M5VBKKTCQ1C1W6DRS): `find` returns only the matching paragraphs and list items, `show` one section or an outline. Reads through prep are also where J1-style filtering or size caps can apply later. Done.
- **G8. Model choice for subagents.** Wide surveys on a smaller model, knowledge lookups across several entries on the smallest; a single lookup stays inline, because a subagent starts with ~13k tokens of fixed prompt. The `prep prime` hint carries this, since the skill loads only when invoked and subagents load neither (01M4DN4NNBCW8YAVB2YP86R0RG, 01M4DNNQJZZY7RJWCKCMH5X5S4). Done.

## First measurements: fathom audits

Four runs on 2026-10-08 of the same prompt in fathom (a Rust project of nine crates and targets, with a fresh knowledge base of about 100 entries): "spawn explore subagents to check if our new prep knowledge base and issues are complete". fathom's state changed between runs (issues and entries from the earlier ones), so the comparison is rough.

| | Run 1 `704a70a0` | Run 2 `3b7d87a5` | Run 3 `a2aed92f` | Run 4 `174fa6b3` |
| --- | --- | --- | --- | --- |
| Setup | before G7 | G7 binary; old skill and ledger still loaded (`/clear`, no restart) | G7, G8 hint, fresh session | hint with "never the files", "unchanged", bounded output |
| Subagents | 5 × Opus | 5 × Opus | 4 × Sonnet | 4 × Sonnet |
| Main loop tokens | 1.51M | 0.60M | 0.65M | 0.52M |
| Subagent tokens | 8.16M | 10.67M | 4.39M | 1.31M |
| Subagent requests | 128 | 149 | 72 | 38 |
| Tool output read back | ~196k | ~207k | ~159k | ~74k |
| of which knowledge (files, or `prep knowledge` in run 4) | ~45k | ~64k | ~43k | ~41k (25 calls: 20 find, 17 show, 6 list; 1 direct file read) |
| Final main context | 105k | 78k | 74k | 62k |
| Cost | $7.32 | $7.50 | $3.14 | $1.60 |

What they showed:

- **Subagents are right for wide surveys.** The main context stayed at 74–105k while subagents read 160–200k tokens; inline, the main context would have needed compaction.
- **Model choice saved the most.** Run 3 cost 57% less than run 1, from the smaller model and fewer requests. Narrower reads contributed little. Whether the Sonnet reports were as thorough as the Opus ones has not been checked.
- **Rules must be in the briefing.** The main agent never invoked the skill in runs 2 and 3; only the `prep prime` hint reached it, and subagents see only what their prompt says. In run 3 the main agent passed the knowledge rule on but softened it to "or the files directly", and the subagents read the files. The hint now says "never the files" and "unchanged".
- **Knowledge reads stayed near 27% of tool output** (K-share above the 25% line) in an audit, where comparing whole entries with old docs is the task. Exact-identifier `grep`s returned 2–3k each and are no waste; bulk `cat` loops over entries and read-back saved output files (two of ~27k characters in run 3, three totalling ~134k characters in run 1) are.
- **Code reads are now the largest share** (~67k in run 3).
- **Run 4: the rules held once worded strictly.** The main agent quoted the hint into every subagent prompt as "passed on unchanged"; the subagents read knowledge only through `prep knowledge`, read back no saved output, and needed 38 requests in all. Run 4 cost 78% less than run 1 and used 81% fewer tokens. Part of that is a smaller task: the main agent told the subagents that open issues already cover known gaps, and their reports were shorter (3.8–5.2k characters against 5–10k). Knowledge volume stayed about the same (~41k); what fell was code and doc reading, requests, and the amplification of both.
- **Tooling pitfalls found:** `prep setup --refresh` uninstalled token-ledger with the marketplace (fixed in 01M4DMXST6T9VKMVF46H9DFR40), and `/clear` keeps plugins and skill text from before an install, so a measured run needs a fresh Claude Code start. The ledger classified `prep knowledge` reads as `prep-cli` until 01M4DP4F2ZF38TCSDVHCFC7PCP; for run 4, add them back to K-share.

Still open: measure ordinary issue work (enrichment, implementation), where narrow knowledge reads should matter more than in audits, and check report quality across models.

## Next steps

1. Record ordinary prep work for one to two weeks (the plugins are installed; start a fresh Claude Code session after `just install`).
2. Compute the metrics above from `~/.claude/token-ledger/*.jsonl`. Claude can do this when given the files.
3. Read the outcome table, pick the matching action, and record the decision in the knowledge base. If the action is a prototype, open a prep issue for it.

## Sources

- [jev-use package](https://pi.dev/packages/jev-use?name=jev): its README, `docs/reference.md` and skill, read from npm `jev-use@0.8.0`
- [TypeSafe Jev explained (eesel.ai)](https://www.eesel.ai/blog/typesafe-jev)
- [Jev on AIMLAPI](https://aimlapi.com/models/typesafe-jev-latest)
- [Portkey TypeSafe integration](https://portkey.ai/docs/integrations/llms/typesafe)
- [TypeSafe research note (rywalker.com)](https://rywalker.com/research/typesafe)
- [Jev on ToolWorthy](https://www.toolworthy.ai/tool/jev)
