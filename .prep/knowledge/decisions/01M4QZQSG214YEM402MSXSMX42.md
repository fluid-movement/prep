---
type: decision
title: "Fresh config directory with passed-through auth"
description: "Fresh config directory with passed-through auth"
date: "2026-10-12"
issue: 01M4QKVBFKQKCQP6N42BG2K8Z4
---

A private CLAUDE_CONFIG_DIR is the only sure way to keep the user's hooks, skills, memory and plugins out; it costs the login, so auth comes from ANTHROPIC_API_KEY or CLAUDE_CODE_OAUTH_TOKEN. Alternatives: --bare (no hooks or skills, which are what we test) and --setting-sources project,local (user skills and memory still load).
