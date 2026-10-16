---
title: 'Step back: refine or remove each part of prep'
kind: manual
tags:
  - rethink
priority: high
---

Step back and look at every step and feature of prep, then refine it or throw it away. The 0.3.0 work showed how easily agentic development builds too much too quickly: the lifecycle orchestration epic doubled the cost of finishing an issue while its evidence gate accepted false reports (see the review of the work since 9c61698). Prep should be an excellent tool, not one that drowns in its own output.

Each child is one area of prep and a decision issue. For each area we settle: what problem it solves and for whom, what evidence says it works, what it costs (agent tokens and commands, code, records, attention), and the verdict: keep, refine or remove.

Process, in order:
1. Frame: this issue and one decision issue per area.
2. The user writes down their thoughts.
3. Discussion per area, enriching each decision issue with findings and the decision.
4. A review round over all decisions together: do they drive prep in the right direction as a whole?
5. Only after that review, create actionable issues from the decisions.

No implementation issues are created and no code changes are made for this work until step 5. Agents do not run ahead of the discussion: one area at a time, led by the user.

## Open questions
