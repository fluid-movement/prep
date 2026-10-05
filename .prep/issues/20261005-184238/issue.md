---
title: Install the Claude Code integration in user projects
kind: code
depends_on:
  - 20261005-152623
---

Users of prep get the Claude Code integration in their own projects without copying files by hand: the prep skill, the SessionStart hook that runs prep prime, and the /prep-status, /prep-next and /prep-guide slash commands. Today these live only in this repository's .claude directory. Installing must be ergonomic, keep the skill in step with the installed binary, and leave a project's own Claude Code settings intact.

## Open questions

- Ship as a Claude Code plugin through a marketplace, have prep init (or a prep command) write the files into the project, or both?
- How does the plugin detect a binary that is too old or too new for its skill and commands?

## Open questions
