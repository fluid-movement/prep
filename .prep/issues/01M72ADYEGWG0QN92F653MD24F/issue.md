---
title: 'prep activity report: what prep costs and whether its rules are followed, from recorded sessions'
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
depends_on:
  - 01M72AV3TQPT1H6GQ6RY46R17T
tags:
  - usage
---

`prep activity report [--since <date>]` summarizes the recorded sessions of a project from the activity stream and its archives, so that a change to hooks or guidance can be judged against real use without writing transcript scripts.

Why: the usage research (01M5FS7KX63E8KZARSXXAS44C4) was done twice in four days with throwaway scripts over harness transcripts, and its second run overturned conclusions of the first (findings 01M6V6R1BBWFSACRAAZX1GPAGP and 01M72A9TBJ5JQXT2VGT15N2HW1). It also tested two proposals by replaying recorded sessions, without a live eval run.

It ships as a command, not as a script kept with the evals (decided 2026-11-10, the user having left the choice to the agent): a prep user needs the same numbers to see whether prep pays off in their project, the stream is per project, and the TUI's Usage view already aggregates part of it.

The report gives, per period:

- hooks: Stop questions asked and skipped, and how many were followed by a decision or finding before the next prompt, by number of changed files; before-tool denials; edit-time decision notices; hook errors;
- rule following: sessions and turns that edited code with no issue read, created or changed earlier in the session; guides followed by a show of the same issue; knowledge lookups before the first code edit of a turn;
- hand-off: briefs written, hand-backs accepted and refused (by reason), and checks that failed, once those commands exist;
- commands: counts and failures by command and error code;
- cost, as far as the stream has it: total from request events, and what share of them the harness reported. Where request events are missing (subagents, harnesses that report none) the report says so instead of estimating. Closing that gap is research issue 01M74VB2DVCENFEXXT7E1CX54A; the report uses whatever cost events the stream has when it is built.

The rule-following figures are given per session and per issue as well as per turn: per turn alone they look worse than they are, because an issue read in one turn is worked on in the following ones (finding 01M72A9TBJ5JQXT2VGT15N2HW1). The guide-then-show and lookup-before-edit figures describe the flow before the guide and edit-time changes and are kept to show their effect.

A turn runs from one prompt hook event to the next; the prompt hook records an event for every prompt, sent or skipped. Text by default, `--json` for the same numbers.

## Open questions
