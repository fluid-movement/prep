---
title: just install keeps the development tooling current
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - install
---

When developing prep, `just install` installs the binary and also refreshes the harness integrations from the checkout (`PREP_PLUGIN_SOURCE=<checkout> prep setup --refresh`), so the Claude Code plugin always matches the code being developed. Release users are unaffected: install.sh runs prep setup and prep update refreshes the plugin pin.

## Open questions
