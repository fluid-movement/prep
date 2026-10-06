---
type: decision
title: Out of scope
description: Topics deliberately left out of prep; each can be added later without changing the model.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# Out of scope

Applies when a request touches one of these topics.

- GitHub integration: none; completion evidence is a plain commit reference.
- Cross-project view: single project only.
- Orchestrating implementation: prep records work but never runs agents.
- Approval or permission system: harness permissions handle it.
- Focus or active goal, goal ordering: none for now. Priorities are in scope since 20261006-090120: fixed levels per issue, no ranking between issues.
- How work inside a step happens: the user's and agent's choice.
- User-level config beyond personal choices: `~/.config/prep` holds only the user's choices, such as integrated harnesses ([Harness setup](/components/setup.md)); project behavior stays in the project config.
- Hosted or database storage: possible through the storage port, markdown only for now.
- Human-facing documentation: can be generated from the knowledge base later.
- Knowledge editing, visualization, publishing: OKF ecosystem tools and `$EDITOR`.
