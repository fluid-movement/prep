---
title: Can Jev classify whether a tool call should be a prep call
kind: research
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
depends_on:
  - 01M4J8780KZRKWHJR4NN8AD85M
tags:
  - adoption
---

Test whether Jev (TypeSafe's judgment model, see `TOKEN-EFFICIENCY.md`) can decide, from a pending tool call and its session context, whether the call should have been a prep call instead (for example a Read of a knowledge file, a grep over `.prep`, a hand edit of an issue record), and which prep command fits. A Claude Code PreToolUse hook could then redirect or annotate the call without the LLM spending tokens on the judgment.

Build a labelled set from the ledger analysis (01M4J8780KZRKWHJR4NN8AD85M and its findings: real tool calls marked "should be prep" or not), measure Jev's accuracy, confidence calibration and latency against it, and compare with plain rules in Go (path and command patterns). Also look at where else in the workflow a typed judgment could help (choosing the next step, picking knowledge entries, triaging drift). Findings recommend whether hooks use Jev, rules, or both, and record the decision in the knowledge base.

## Open questions
