---
type: decision
title: Research handback carries criteria and completes when met
description: Research reports without --check write findings, apply criterion verdicts and complete only when all are met and Problems is empty.
date: "2026-11-11"
issue: 01M72CWZ12ESB9D853ENN2NQ67
supersedes:
  - 01M75H9X76QTXSF3P8CXXXB9AS
---

The earlier checker-only decision contradicted the issue requirement and acceptance criterion 3: a research report writes findings and completes the issue. Research completion gates require checked criteria, so its report must carry a Criteria section with a verdict per criterion, check the met ones and complete when all are met and Problems is empty. Otherwise it logs unmet criteria and problems and leaves the issue in progress. Code implementer reports remain ready-for-check reports; checker reports still carry criteria. Preserve internal/domain/fsm.go gates.
