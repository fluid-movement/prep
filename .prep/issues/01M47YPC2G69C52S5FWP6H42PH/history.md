- 2026-10-06T06:48:43Z edited by claude-code/2.1.289: tags

- 2026-10-06T12:04:09Z edited by human:azaharias: parent

- 2026-10-06T12:04:23Z edited by human:azaharias: parent

- 2026-10-06T13:58:01Z edited by claude-code/2.1.291: requirement

- 2026-10-06T14:22:26Z claude-code/cloud: prep list --text matched titles only, so the skip rule (find a candidate's Source: line) could not work; --text now also matches the requirement. This widens --text in views and the TUI filter too.

- 2026-10-06T14:22:26Z claude-code/cloud: The import issue is open like the bootstrap issues: the user defines it before the agent starts. The context warns that a path:line source can be the prefix of a longer one, so the agent checks the found Source: line exactly.
