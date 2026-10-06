- 2026-10-05: Hook command tested in a shell with and without prep on PATH; both exit 0, silent without the binary.
- 2026-10-05: Claude Code picked up the three commands mid-session; /prep-status, /prep-next and /prep-guide 20261005-152620 ran and rendered as intended. The SessionStart hook itself fires from the next session on.
- 2026-10-05: Dropped the /prep-status instruction to show children under indented parents; prep list does not indent, so parent grouping is inferred from the issue files.
- 2026-10-05: Filed 20261005-184238 for installing the integration in user projects, blocked on the install decision.

- 2026-10-06T06:48:44Z edited by claude-code/2.1.289: tags
