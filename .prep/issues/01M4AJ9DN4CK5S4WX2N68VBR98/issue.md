---
title: 'Remove commit mode: prep only stages its writes'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - config
  - git
---

prep no longer commits on its own. `commit_mode` and its `all` value go away: after every write prep stages the files it touched, and the user or agent commits them together with the code, as the default already does. Committing after every prep action would produce a flood of tiny commits, and the records already carry who and when.

- `commit_mode` is removed from the configuration, `DefaultConfig`, the TUI settings screen and the documentation (README, skill, knowledge).
- A config file that still sets `commit_mode` is reported by `prep check` like any unknown key; there is no migration, since prep has no users outside this repository. This repository and the test fixtures under `testdata/` are converted.
- The decision in "Default commit mode" (01M46ARS5RNN4X2TMVHAKWA3KW) is superseded: staging is the only behavior.

## Open questions
