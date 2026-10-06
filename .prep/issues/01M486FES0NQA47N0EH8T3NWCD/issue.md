---
title: 'Harness integrations: prep setup and the Claude Code plugin'
kind: code
parent: 01M48FAKAG2X998F2ZHE1MHVTC
tags:
  - install
  - integration
---

The binary is only half of prep: the skill, the session-start hook and the commands are what make an agent understand and use it. Installing prep therefore sets up the harnesses the user works with, and keeps them in step with the binary.

- prep setup provides a surface any harness can plug into; Claude Code is the first implementation. Other harnesses, such as Pi, come as their own issues.
- Integrations are installed per user, so every session on the machine knows prep; repositories carry no harness files.
- A user-level configuration in ~/.config/prep (XDG_CONFIG_HOME) records the user's choices, starting with the integrated harnesses. What is actually installed is asked from the harness, not stored.
- install.sh runs prep setup after installing the binary when a terminal is attached: it detects the harnesses on the machine, preselects them, asks to confirm and installs. prep update refreshes the integrations automatically, so binary and integrations always carry the same version.
- For Claude Code the integration is a plugin served from the prep repository as its own marketplace, holding the skill, a SessionStart hook and the status, next and guide commands, structured so the TUI mod can join later.

## Open questions
