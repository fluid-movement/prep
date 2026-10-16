## D1: An issue is clear when an agent without the conversation could do it
date: 2026-10-16

User agreed, 2026-10-16. The written issue is the handoff: the next session, a subagent or another harness has only the records. Intended use: a heavy main agent holds context, talks with the user, defines and readies many issues and checks results; cheaper subagents implement ready issues without polluting the main thread. Prep makes issues handoff-ready; the harness runs the subagents, prep does not schedule.

## D2: Steps: settle requirement, document how, implement, check, done
date: 2026-10-16

User decision, 2026-10-16. Define and ready stay separate steps. Define: the requirement is settled, and the agent checks with the user that it is what they want. Ready: the how is documented, so any agent can take the issue. Then implementation, then the result is checked, then done. Kinds stay: they are a filter and say what the outcome of an issue is.

## D3: Prep owns the gates, not the path
date: 2026-10-16

User direction, 2026-10-16, restating design principle 1. Prep defines what must exist before an issue can move on; how the agent and human get there is theirs. Guidance (skill, guide) suggests a path without enforcing it. A gate is only worth having if it checks something the next reader needs and that a check can meaningfully catch; whatever prep cannot check stays guidance, and quality of what passes a gate is measured by evals, not gates. The how is discussed with the user where there is a real choice, as guidance, not as a required confirmation.

## D4: No separate checked state
date: 2026-10-16

User agreed, 2026-10-16. Implementers do not complete; they log that the work is implemented and ready for checking, and stop. Whoever checks (main agent, with the user when needed) completes. The claim and log show work waiting for a check. Promote to a recorded state only if work gets lost in practice.
