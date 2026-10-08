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
  - every tool call, with what it touched, its prep area (`knowledge`, `issue`, `prep-cli`, `code`), the prep subcommand, and how much text it returned;
  - context fill and cost at the end of each turn.

  The row format is in its README. To load it from a checkout: `claude --plugin-dir plugins/token-ledger`.
- **The prep pane's Usage tab** (`/prep:pane usage`) shows the current session's figures as they accumulate.

Record ordinary prep work, unchanged, for one to two weeks. Aim for at least 10 sessions that each take one issue through enrichment or implementation.

## What we compute

| Metric | Definition |
| --- | --- |
| **K-share** | Knowledge-entry tokens read back ÷ all tool-output tokens, per prep session |
| **K-per-issue** | Knowledge tokens read in a session that works on one issue |
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

## Next steps

1. Merge or install the two plugins, and record ordinary prep work for one to two weeks.
2. Compute the metrics above from `~/.claude/token-ledger/*.jsonl`. Claude can do this when given the files.
3. Read the outcome table, pick the matching action, and record the decision in the knowledge base. If the action is a prototype, open a prep issue for it.

## Sources

- [jev-use package](https://pi.dev/packages/jev-use?name=jev): its README, `docs/reference.md` and skill, read from npm `jev-use@0.8.0`
- [TypeSafe Jev explained (eesel.ai)](https://www.eesel.ai/blog/typesafe-jev)
- [Jev on AIMLAPI](https://aimlapi.com/models/typesafe-jev-latest)
- [Portkey TypeSafe integration](https://portkey.ai/docs/integrations/llms/typesafe)
- [TypeSafe research note (rywalker.com)](https://rywalker.com/research/typesafe)
- [Jev on ToolWorthy](https://www.toolworthy.ai/tool/jev)
