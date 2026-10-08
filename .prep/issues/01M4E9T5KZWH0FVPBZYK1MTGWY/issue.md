---
title: 'CLI and integrations: -- in arguments, activity add, the bridge''s actor, setup and update rollback, watcher errors'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - cli
  - integration
---

Found in the 0.2.0 codebase audit.

- Global flags were stripped anywhere in the arguments, so a log text of `--by x` failed and `-json=false` turned JSON on; `--` did not protect text (and `parse` lost a `--` after the first positional).
- `prep activity add`, which the bridge runs after every tool call, loaded the whole tree just to find the project.
- Events clipped only verb and target, so long op, issue, actor or session fields could exceed the atomic append size.
- The bridge left prep commands to prep, but prep records reads only for a named actor, so an agent running prep without `--by` left no reads.
- `prep setup` removed the marketplace before adding the new pin; a failed add left the user with no marketplace and no plugins.
- On Windows a failed rename during `prep update` left no `prep.exe`; install.sh reported a URL as the version when there is no release.
- The watcher discarded fsnotify errors (missed changes went unreported) and sent from timer goroutines, contrary to its own invariant.

## Open questions
