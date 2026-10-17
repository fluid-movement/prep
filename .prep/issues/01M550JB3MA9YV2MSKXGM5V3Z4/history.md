- 2026-10-17T13:28:20Z codex: Implemented, ready for check: Bash record redirects now parse command paths and redirections, exempt prep/git and quoted or heredoc prose, and preserve direct read/write denials with regression tests and updated hooks knowledge.

- 2026-10-17T13:30:52Z claude-code: Checker: command wrappers (sudo, env, timeout, nice, xargs, ...) are unwrapped so a wrapped read or write of a record is still redirected; tests added.
