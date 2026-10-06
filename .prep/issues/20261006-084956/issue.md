---
title: 'prep setup: install harness integrations'
kind: code
tags:
  - install
  - integration
---

The binary is only half of prep: the skill, the session-start hook and the commands are what make an agent understand and use it. Installing prep therefore includes setting up the harnesses the user works with. After install.sh has installed the binary it runs prep setup, which asks which harnesses the user uses (Claude Code, Pi, others as they are supported) and installs their integrations: skill, hooks, commands or plugin. prep setup can be run again at any time, takes the harness list as an argument for non-interactive installs, and prep update refreshes installed integrations so they always match the binary's version. The setup logic lives in the binary, so it is testable and works where install.sh does not (Windows); install.sh stays a thin downloader. Prompts read from the terminal, because curl | sh leaves no keyboard on stdin.

## Open questions

- Which parts are installed per user (every session knows prep; the hook must stay silent outside a prep project) and which per project (committed to the repository, as this repository does today)?
- For Claude Code: a plugin through a marketplace, files written into ~/.claude, or both?
- How does an integration detect that the installed binary is too old or too new for it?
