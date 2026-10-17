## D1: User input: document decisions, readers are agents, relevance at the right time
date: 2026-10-16

User, 2026-10-16. Only agents read knowledge. Whether the current base helps is unknown and there is no way of telling yet; it matters more as a codebase grows, but only if the relevant information is served at the right time, which is a hard, open problem. Documenting what the system does drifts as the system changes; decisions are timeless: a decision between a and b can later be overridden, and the documented reasoning for a lets us make a better call for b. Areas to settle: what we document, who documents, who reads. For reading: decision models (such as Jev) and local models may become good at judging whether entries are relevant to a prompt, which favors a large orchestrator that hands off context-heavy work and receives just what it needs at the right time.

## D2: Knowledge is decision records, unified with issue decisions (trial)
date: 2026-10-16

User agreed to try, 2026-10-16. Knowledge consists of decision records: context, options, the choice and why; never edited, only superseded. A decision made in an issue (prep decide) is a knowledge entry linked to that issue, so decisions are findable outside the 220 issues they are buried in today. Descriptive entries shrink to what can be phrased as a decision or constraint; conventions are decisions too. Decisions are scoped to code paths so hooks can surface them when an agent touches that code; each has a crisp one-line description so a cheap or local model can triage the index later. The orchestrator links relevant decisions when it readies an issue. Value is measured by an eval: does an agent respect a recorded decision or explicitly supersede it, rather than silently violate it.

## D3: Project overview lives in README or AGENTS.md, not knowledge
date: 2026-10-16

User decision, 2026-10-16. What the project is and where things live belongs in the project's README or AGENTS.md, which harnesses already load; the knowledge base holds decisions only.

## D4: No knowledge gate at completion; nudge instead
date: 2026-10-16

Agent's choice at the user's request, 2026-10-16. For a gate: it forces a moment to ask whether anything was decided, and agents were seen leaving knowledge behind (01M4QY36MC01A53N0MM0S55FCK). Against: prep cannot tell a considered --no-impact from a reflexive one, so it fails the gate test, and costs a command and boilerplate on every issue. With decisions recorded as they are made (prep decide is the knowledge write), there is nothing left to do at completion. The finish hook and guide nudge instead: when a turn changed code, ask whether a decision was made and record it. The decision eval shows whether the nudge suffices; a gate comes back only if it does not.

## D5: Issues are the work; knowledge is what was decided and learned
date: 2026-10-16

User proposal with agent agreement, 2026-10-16. Issues hold the work: requirement, how, criteria, log and state; they change until done. Knowledge holds what was decided and what was learned: decision records and findings, dated and never changed, only superseded by newer records. Research findings move from findings.md in the issue to finding records in the knowledge base, linked to the issue. Kind outcomes follow: a decision issue produces a decision record, a research issue produces a finding record, which makes their completion gates checkable (a linked record exists).

## D6: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16/17, are carried into release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) by 01M52ZR294ZMVPKF7N2KPPKD9Q, 01M52ZRB3VY0MSZ0HZWJ2SEC24, 01M52ZRE24EMGBHX3SKV62436G, 01M52ZRSZFBMWZQJPAXVHXD7M1. Open questions raised here were settled by the release's decision issues.
