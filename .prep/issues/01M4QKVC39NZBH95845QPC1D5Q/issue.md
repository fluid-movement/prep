---
title: 'Assertions on tool use: used, not used, order, count, arguments'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVBMEVANE3EBQM2V2VKR8
  - 01M4QKVBSB5FWC25MM7CB8RBNM
tags:
  - testing
---

## Requirement

Deterministic checks on the normalized transcript. A case lists them under `expect.tools`; each passes or fails with a one-line reason naming the step that decided it. Kinds:

- `ran`: a prep command (by name, optional arguments and project) was run, optionally at least/at most N times.
- `not_ran`: a command, or a tool (Read, Grep, Edit...) used on a path pattern, never happened. Example: no Read of `.prep/**`.
- `before`: A happened before B (for example `guide` before `edit`).
- `max_calls`: tool calls overall, or of one tool, stay at or under a number.
- `no_errors`: no tool call failed, or only the listed ones.
- `args`: a run of a command had given flags or values (regular expression on the arguments).
- `final_mentions`: the final message contains text.

Matching works on parsed prep commands, so `prep --project api list` and `cd x && prep list --project api` are the same command.

## Open questions
