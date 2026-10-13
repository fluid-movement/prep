---
title: TUI Inbox and bundled completion review
kind: code
parent: 01M4TST3C4ZSZDKA5WKBA3ND4H
depends_on:
  - 01M4TSWXEPBGMQG271CZJBBZJP
tags:
  - tui
---

Add Inbox as a top-level TUI screen alongside Issues, Agent, and Knowledge. Prioritize completion reviews, ready sign-offs, open questions, stale requirements or completion policies, and dependency uncertainty; keep ordinary backlog in Issues. Completed work enters Inbox immediately with a non-blocking notification; do not steal focus. The review screen bundles requirements, code changes, acceptance and DoD evidence, knowledge changes, and exceptions. Approval is its only review-verdict action; provide easy issue-ID copying so feedback and requests for changes are discussed with the agent. Leaving the screen without approval keeps review pending. Reopening occurs explicitly through the agent after human instruction, not through a reject/request-changes UI. Reuse useful behavior and tests from 01M4KHRQ02GQQ7F7XRY7KE2BR8 and 01M4KHRQ4D0GKRS5GX6CNWRFVR, adapting machine-local review state to the new versioned contract.

## Open questions
