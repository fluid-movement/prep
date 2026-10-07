---
title: 'Skill: prep only stages; commit .prep changes with the code'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
depends_on:
  - 01M4AVPH9A56WJ4JFMYZT476H0
tags:
  - workflow
---

Since commit mode was removed, prep stages the `.prep` files it writes and never commits; the skill does not say so, so an agent may leave records uncommitted or commit them apart from the code. The skill states that prep stages its writes and that the agent commits them together with the code they describe, and how completion evidence works (see the completion issue). Its list of read commands also gains `flags`, `views` and `theme list`, and writes gain `theme new`.

## Open questions
