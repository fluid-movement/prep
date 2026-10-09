---
type: component
title: TUI knowledge screen
description: 'internal/tui knowledge screen: entries with attention marks, the entry view, filters and links between issues and entries.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-07T08:11:57Z
scope:
  - internal/tui/knowledge.go
confirmed_commit: ad855be3907fbc5ab7df8b7e06529318670b1a9b
---

# TUI knowledge screen

Applies when changing how `prep tui` shows the knowledge base.

- (`knowledge.go`, 01M46ARHBR3W5H5HDEHS0RB37T): read-only, since agents write knowledge for agents and nothing in the TUI edits it. `b` opens or leaves it from any screen. Left: the entries (`Tree.QueryKnowledge`) as rows of type, status, an attention mark (`!` in the warning tone) and title. Right: the entry in its own viewport (`shownEntry`: only while the list shows it, so a filter that empties the list shows "Nothing selected" and keeps the selection for when it is cleared) (`knowState.vp`) with a meta line (type, status, path), description, scope and confirmed commit, the check's findings, "Changed by" (`Tree.EntryIssues`), then the body under its own heading. Keys (`knowledgeBindings`): arrows or j/k select, enter scrolls the entry, `f` filters (`/` too, unlisted as on the issue screen; `ParseKnowledgeFilter`: `--type`, `--status`, `--scope`, words; the footer lists those flags while typing), `!` shows only entries needing attention, `o` goes to an issue that changed the entry, backspace returns to where the person came from (the back stack of [TUI navigation](/components/tui-navigation.md)), esc clears the filter or leaves; no second-tier views yet, the bar's second line shows the entry count (01M4EA4RZ6VDWCR6XR6DCGAA8T researches which views fit). Attention comes from `Options.Check` (K003 broken link, K004 size, K005 drift, K006 unknown commit), run when the screen opens and on reloads while it is open; the title says "checking" until it returns. The issue detail lists the issue's entries (`Tree.IssueEntries`: context links, and once resolved the documented entries) as `≡ knows <type> <title>` links; `o` opens one in the knowledge screen.
- Part of the [TUI](/components/tui.md); entries come from the [OKF store](/components/okf-store.md).
