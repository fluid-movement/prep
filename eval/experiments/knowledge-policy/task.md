# Shared Codex task prompt

The prompt and contract context are identical in all four conditions. Only
advisory knowledge entries differ; a shared minimal `/overview.md` bootstrap
scaffold and the same Codex prep skill are present in every fixture.

> the headline on the player page gets cut off after 80 characters, i think
> the database column is too small. fix it.
>
> Use the already-tracked issue. Follow the project's current prep workflow
> and completion gates, including truthful evidence and an explicit
> knowledge review report. Do not invent approval or claim checks you did not
> perform. Read and follow `.agents/skills/prep/SKILL.md`, which is identical
> in every condition. Treat knowledge entries as guidance and inspect the
> current code before changing it. Run a relevant check and report what you
> actually ran.

The second paragraph is common policy/gate context, not condition-specific
knowledge. The stale condition intentionally contradicts the repository at
task start; the prompt asks the agent to verify it against code.
