---
title: 'Per-user project configuration: gitignored config.yaml with a committed config.yaml.dist'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - config
  - tui
---

Each person working in a prep project has their own `.prep/config.yaml`, so views, commit mode and TUI settings such as the theme are personal and never conflict in git. The project commits `.prep/config.yaml.dist` as the shared default; `.prep/config.yaml` is gitignored. prep reads `.prep/config.yaml` when it exists and falls back to `.prep/config.yaml.dist` otherwise, so a fresh clone works without any setup and a user creates their own config by copying the dist file.

- `prep init` writes `config.yaml.dist` (today's default content) and a `.prep/.gitignore` entry for `config.yaml`.
- Every reader of the project configuration (`prep views`, `prep list --view`, commit mode, the TUI) uses the same lookup: own config, then dist.
- Writes that change the configuration (for example saving a view from the TUI) go to `.prep/config.yaml`, creating it from the dist file first.
- `prep check` validates whichever file is in effect and accepts `config.yaml.dist` as a known project file.
- This repository is converted: its config.yaml becomes config.yaml.dist and config.yaml is ignored.

## Open questions

- 0.1.0 is released: how do existing projects with a committed `.prep/config.yaml` move over — a `prep fix` or `prep migrate` step that renames it to config.yaml.dist and adds the ignore entry, or a note in the release notes?
- Should `commit_mode` stay per-user, or is it a project rule that belongs only in the dist file?
