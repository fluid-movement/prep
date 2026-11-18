---
title: 'Cloud session spike: one issue from push to merge'
kind: research
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

Run one real prep issue through a cloud session from push to merge, to learn what breaks before `prep handoff` is built. Use the cloud environment from the decision on the environment setup script (prep v0.3.0 or a commit build).

Steps: push main; pick a small ready code issue; start `claude --cloud` with the prompt "run prep guide <id>, implement it, log ready for check and stop" (or the text from `prep brief <id>`); watch the session; then bring the result back: fetch the session's branch, merge it, and read what arrived in `.prep/`.

Questions to answer, each as a finding: which branch the session pushes and how the user finds it; whether the activity and work-log records from the cloud session merge cleanly into main (index files, activity stream); whether `prep brief`/`prep handback` fit a cloud session better than `prep guide`; whether two sessions on two issues in parallel conflict; whether PreToolUse hooks fire in the cloud, and what the blank `request` event and the `hook skipped` event in the activity stream mean; what a session costs against usage compared with a local subagent.

## Open questions
