---
type: decision
title: "TUI only: issue screens, Agent screen, decision lookup; desktop and extra themes go"
description: "TUI only: issue screens, Agent screen, decision lookup; desktop and extra themes go"
date: "2026-10-16"
issue: 01M4ZCMBPJ5FM06F9H70XMPE9H
---

User decisions, 2026-10-16. TUI only: delete the desktop app (about 6,600 lines of Swift) now that the Inbox it was to hand over is removed; its open issues (01M4JVP89F3PGDHYETHEPS3YBB, 01M4JVPRARK7X2MB05BNSD8WAA, 01M4JVPRRGXSKRB2AY8PJSHSP3, 01M4JX2N231RDG9KEHDBC1ZRE1) are dropped and 01M4TSY329F505HDC22HFYPD42 becomes the deletion, when actionable work is created. Keep the issue screens (list, views, filters, tree, detail), simple editing, mouse and the Agent screen including its Usage tab (useful, cheap, in its own tab). Keep the Knowledge screen as a read-only lookup of decision and finding records, showing the links from an issue to the decisions made in it. Themes: keep one dark and one light theme; remove the other built-in themes, custom themes, prep theme new and the theme preview. Remove the Inbox. Fix the live refresh bug (01M4WYKSGGDM7R3FSKE8451X2X).
