---
title: 'Initial suite: one set of cases per CLI area'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVBYFPQT50HP41EHZ6F4X
  - 01M4QKVC39NZBH95845QPC1D5Q
  - 01M4QKVC87EPF7DT9AN3NTGF6B
  - 01M4QKVBAARQ1BSMG1X570F9HK
tags:
  - testing
---

## Requirement

The first real suite, plus the commands that run it. `prep eval run` takes case selectors (area, name, tag), variants, models and N, builds the matrix, runs it and prints the report; `prep eval list` lists cases; `prep eval doctor` runs the isolation leak check and a harness smoke run.

Cases cover each part of the CLI through the flows the skill and hooks exist to protect, with a few cases per area: session start and picking work (`prime`, `next`), defining (requirement, open questions, `define`), enriching (context, decisions, criteria, `ready`), implementing (`claim`, work log, `complete` with docs or no-impact), creating and editing issues (parent, dependencies, tags, priority), reading knowledge through prep rather than files, writing and confirming knowledge entries, links and cross-project work (`--project`, cross dependencies, knowledge links), queries and views, and dropping and releasing. Each case has tool-use and outcome assertions; a rubric only where text quality matters. A subset tagged `smoke` runs in under a few dollars.

## Open questions
