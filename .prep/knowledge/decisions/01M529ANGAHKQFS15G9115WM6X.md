---
type: decision
title: "Remove human approval records and the Inbox"
description: "Remove human approval records and the Inbox"
date: "2026-10-16"
issue: 01M4ZCM3ZJDTM3JRHS2N1NMF5H
---

User decision, 2026-10-16. Done is settled in the session between agent and human; the human needs no approval records in prep and no Inbox. Either the agent signals that it needs the human, or the human asks for work and the agent says what is needed; the issue list shows each state. Remove review approve, resolution reviews, approval prerequisites on dependencies, dependency override, the Inbox screen and its completion bundle (about 1,500 lines across domain, CLI and TUI). This also restores design principle 2: prep has no approval mechanism of its own.
