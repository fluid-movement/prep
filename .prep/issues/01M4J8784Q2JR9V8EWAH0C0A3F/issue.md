---
title: Can Jev classify whether a tool call should be a prep call
kind: research
depends_on:
  - 01M4J8780KZRKWHJR4NN8AD85M
tags:
  - adoption
---

## Requirement

Test whether Jev (TypeSafe judgment model, see `TOKEN-EFFICIENCY.md`) can decide, from a pending tool call and its session context, whether the call should have been a prep call instead (for example a Read of a knowledge file, a grep over `.prep`, a hand edit of an issue record), and which prep command fits. A Claude Code PreToolUse hook could then redirect or annotate the call without the LLM spending tokens on the judgment.

Build a labelled set from the ledger analysis (01M4J8780KZRKWHJR4NN8AD85M and its findings: real tool calls marked "should be prep" or not), measure Jev accuracy, confidence calibration and latency against it, and compare with plain rules in Go (path and command patterns).

Also evaluate prompt classification for configurable prompt hooks: given a user prompt in a Prep project, can Jev choose whether to inject no Prep context, a concise workflow reminder, or issue-specific context? Extract and validate an explicit issue ID deterministically in Prep; use Jev for the fuzzier intent/context-level decision. Compare Jev with simple rules and with session-start guidance alone. The hook should short-circuit without a Jev request when prompt hooks are disabled or the working directory is not a Prep project. Measure classification quality, confidence, latency, token/runtime cost, and the implications of sending prompt text to an external service. A low-confidence or unavailable result should fall back to a documented minimal behavior.

Use a labelled set for the prompt cases as well as tool calls, including ordinary questions, project work, explicit issue references, and prompts with ambiguous intent. Findings should recommend whether hooks use Jev, a local model, deterministic rules, or a combination, and record the decision in the knowledge base.

Test a local model as a third candidate next to Jev and the rules: a small instruction model (about 3–8B parameters) served on the developer's machine by a local model server such as Ollama. Measure it on the same labelled sets for tool calls and prompts: classification quality, confidence, latency both warm and after the model has been unloaded, memory use, and the setup it adds for a prep user. A local model keeps prompt text on the machine. Like Jev, it is optional, with the same documented fallback when it is not running.

## Open questions

- Should the prompt classifier run for every prompt when enabled, or only after a cheap deterministic prefilter?
- What should happen on low-confidence or failed Jev classification: inject a minimal reminder, skip injection, or use another fallback?
- What disclosure and opt-in should apply when user prompt text is sent to Jev?
