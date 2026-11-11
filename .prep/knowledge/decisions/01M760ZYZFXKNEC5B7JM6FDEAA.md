---
type: decision
title: 'Hand-off report format: role-specific required criteria and research completion'
description: Summary is required; checker and research reports also require Criteria, with every criterion verdict, even when the section is empty.
date: "2026-11-11"
issue: 01M72CWZ12ESB9D853ENN2NQ67
supersedes:
  - 01M75C2YSMW2Y8P3Z5YZBVZKAT
---

Defined once in internal/domain/report.go for brief and handback. Sections are ## Summary, ## Criteria, ## Decisions, ## Findings and ## Problems; empty optional sections may be omitted. Every report needs Summary text. Checker reports require Criteria even with no criteria and cannot contain Decisions or Findings. Code implementers report Summary, Decisions, Findings and Problems without criterion verdicts. Research implementers also require Criteria and at least one finding. Each criterion is <number>. met — <evidence> or <number>. not met — <what is missing>, with a verdict per criterion. A decision or finding uses ### <title> followed by its reason or finding text; an optional Scope: <glob>[, <glob>] line directly below the title sets scope. Heading case, blank lines, verdict punctuation and unknown extra sections are tolerated. Research reports without --check write findings and apply verdicts; all met with Problems empty completes the issue under unchanged gates, otherwise unmet criteria and problems are logged and it stays in progress. Checker reports apply verdicts and complete on the same condition. Code implementers log ready for check instead. This corrects the stale permission to omit an empty Criteria section and aligns research with the issue requirement.
