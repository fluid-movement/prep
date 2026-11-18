---
type: finding
title: 'Correction: zero-usage requests are rejected unless status is failed'
description: 'Corrects the earlier bridge finding: Validate accepts a failed request with no usage.'
date: "2026-11-18"
issue: 01M7Q923PB7QG1D1XZ0075V63M
---

Corrects the finding "Activity bridge: zero usage and lifecycle diagnostics", which says `Event.Validate` rejects request events with no non-zero tokens, context or cost "even when status is set". The final rule is: such a request is rejected unless its status is `failed`, because a failed request legitimately carries no usage (see the decision "Failed requests stay valid without usage; zero-usage requests are rejected"). Findings are immutable, so this record supersedes that sentence.
