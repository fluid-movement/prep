---
type: finding
title: First review by a subagent passed what the main session caught
description: A Sonnet reviewer working from the checker brief passed the change but missed two design weaknesses
date: "2026-10-18"
issue: 01M56ZKEHAQE5C4G8J0NQA0KPX
---

2026-10-18, first use of the review pattern: a Sonnet subagent checked the handoff change from prep guide's checker brief and the diff alone. It verified every criterion correctly, ran the checks and reported one real finding (its own accidental log line). It passed two design weaknesses the main session caught from context: the implement handoff prompt did not tell a handed-off agent that someone else checks, and the check step ended on any later log line. Both were outside the criteria, which the reviewer checked faithfully. Lesson: a review subagent verifies what the issue states; design gaps in the issue itself still need the session that defined it, or criteria that name the behavior.
