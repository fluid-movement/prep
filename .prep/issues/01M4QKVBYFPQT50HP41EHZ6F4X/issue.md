---
title: 'Session runner: run Claude Code headless on a prompt in a fixture'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVBFKQKCQP6N42BG2K8Z4
  - 01M4QKVBMEVANE3EBQM2V2VKR8
  - 01M4QKVBSB5FWC25MM7CB8RBNM
tags:
  - testing
---

## Requirement

The session runner is the deterministic test runner. Given a case, a variant and a model, it builds the fixture workspace, sets up the isolated environment, runs Claude Code headless on the case's prompt from the primary project, enforces the case's limits, and returns a run record: the normalized transcript, the raw stream, the workspace left behind for state assertions, the exit status and timing. It does not judge anything.

A run ends cleanly on completion, and is cut off, with the reason recorded, on timeout, turn limit or budget limit. A harness failure (binary missing, auth missing, crash) is reported as an error of the run, never as a failed assertion. The model is a parameter of the run, so the same case runs unchanged on any model. Running stays on this machine and uses only the isolated environment.

## Open questions
