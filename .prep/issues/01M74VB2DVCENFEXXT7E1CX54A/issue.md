---
title: Why request events are incomplete, and how the stream gets a turn's token use including subagents
kind: research
parent: 01M5FS7KX63E8KZARSXXAS44C4
tags:
  - usage
---

Find out why the activity stream holds far fewer request events than the harness made requests, and decide how prep gets a session's token use completely enough for `prep activity report` to state what prep costs.

Why: the parent's question is cost, and the stream cannot answer it. Daypack's stream held 125 request events for two days in which the transcripts show several hundred requests, and subagent requests are not in it at all, although subagents did most of the implementation there (39 subagents). Both usage findings were therefore made with throwaway scripts over harness transcripts. The stream issue (01M72AV3TQPT1H6GQ6RY46R17T) leaves request events out and the report issue (01M72ADYEGWG0QN92F653MD24F) only states the coverage, so without this issue nobody owns the gap.

Questions to answer:

- Which requests does the Claude Code plugin module report today, and which does it miss (subagents, background tasks, requests after a session ID change, sessions where the module is not loaded)?
- Can the finish hook close the gap instead: do the Stop and SubagentStop hook inputs carry the transcript path (and the subagent's own), and can the binary sum the usage of the turn's requests from it, deduplicated by message ID as the research scripts did, and write one cost event per turn with the agent ID?
- What does Codex offer for the same?
- Which of the two sources should the report use, and should the other go?

The result is a finding with the measured coverage per source, and a code issue for the chosen route, or a decision to leave cost to transcripts and say so in the report.

## Open questions
