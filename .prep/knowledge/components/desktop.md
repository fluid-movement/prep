---
type: component
title: Desktop app
description: 'desktop/: the macOS app (Swift, SwiftUI) over the prep CLI''s JSON — PrepKit, the window model, project switching, building with just desktop, and how agents see it with snapshots.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-10T12:36:46Z
scope:
  - desktop/**
---

# Desktop app

Applies when changing the macOS app in `desktop/`, or CLI output it decodes.

- **Boundary**: the app talks to prep only through the binary: `prep --json --root <dir> …` for reads and writes and one `prep watch --json` per open project. It never reads `.prep/` records; the only filesystem lookup is finding the project root (nearest directory with `.prep`, as `mdstore.Find` does). Any JSON change in the [CLI](/components/cli.md) that renames or drops a field the app decodes breaks it; `swift test` runs the real binary and catches that.
- **Package** (SwiftPM, no Xcode project): `PrepKit` holds the Codable models of the JSON (`Models.swift`), `PrepCLI` (locates the binary through `PREP_BIN`, `PATH`, the usual install directories, then a login shell, since Finder gives apps a minimal PATH; global flags go before the command because prep reads everything after `--` literally), `Watcher` (debounced `prep watch` follower; `stopAll` on quit), `IssueTree`, `Query` (filter text to list flags) and a small block `Markdown` parser (inline markup through `AttributedString`). `PrepDesktop` is the SwiftUI app: `ProjectModel` per window, `ContentView` (split view, toolbar), `IssueDetailView`, `MarkdownView`, `Snapshot`.
- **Actor**: reads pass no `--by`, so the app's reads stay out of the activity stream; writes pass `--by human:<user>`.
- **Issues screen**: search plus filter menus (view, state, priority, kind, tag, more: actionable/blocked/stale, and a subtree chip) built from `prep flags --json` with live counts; `IssueFilter` (PrepKit) turns them into `prep list` flags with the saved view first. Tree or flat list; rows show state symbol, title, stale/blocked, priority bars, child progress ring and short ID; context menus copy, add a child, filter to the subtree, reveal in Finder.
- **Issue detail** (`IssueDetailView`, `Relations`): breadcrumbs of ancestors with sibling stepping, title (double-click renames), a lifecycle card (steps open → done, the next transition as the primary button, unmet gates in plain words with a button that opens the fixing editor), a dependency map (depends-on cards → this issue → blocked cards, joined by arrows; green when done, dashed orange while blocking), a children card (state bar, unresolved first, resolved folded), then cards for requirement, acceptance (checkboxes, add field), context, decisions, findings and the work log timeline. Transitions are confirmed in a sheet with fields (reason, documentation entries or no-impact, commit, note), not raw flags. Editors (`EditorSheet`) write title, requirement, context, findings, criteria and decisions through the matching prep command.
- **Inspector**: kind, priority, parent and tags editable in place (`prep edit`), the timeline from the ULID, baselines, ready, claim and resolution.
- **Knowledge and Agents**: entries by type with search and a rendered body; agent sessions (live within 3 minutes, current issue) and the activity feed by day, model requests behind a toggle.
- **Navigation and keys**: back/forward history (⌘[ ⌘]), ⌘↑ parent, ⌘1–3 screens, ⌘F or / search, ⇧⌘N new issue, ⌥⌘N new child, ⌘E edit requirement, ⇧⌘R rename, ⌘↩ next step, ⌥⌘I inspector, ⇧⌘C copy ID, ⌘R reload.
- **Look**: SF Symbols per state, kind and priority (`Style.swift`), cards on the control background, system accent, light and dark; the app icon is drawn in SwiftUI (`AppIconView`) and `just desktop` renders it into `AppIcon.icns` (`--render-icon`). Settings shows the prep binary in use.
- **Projects**: the toolbar dropdown lists the [project registry](/components/projects.md) merged with projects the app opened itself (older binaries have no registry); the window opens the launch directory (`PrepDesktop <dir>`, `open -a Prep <dir>`), else the last project.
- **Build**: `just desktop` assembles `desktop/build/Prep.app` (Info.plist from `desktop/Resources`, ad-hoc signed); `just desktop-test` runs `swift test` against `./prep`; CI runs both on `macos-15`. Developer ID signing, notarization and a Homebrew cask are not done yet.
- **Seeing the app (agents)**: `PrepDesktop <project> --snapshot out.png [--select <id>] [--tab issues|knowledge|agent] [--entry /path.md] [--size WxH] [--appearance light|dark] [--sheet new|requirement|context|findings|criterion|decision|title|drop|complete|release]` (or `just desktop-snapshot out.png …`) loads a real window off-screen and writes a PNG without Screen Recording permission: `cacheDisplay`, plus compositing for Liquid Glass, which the window server draws (flat stand-ins for glass panels, their content drawn on top, list rows redrawn, toolbar items last; every scroll view is cleared and its visible content drawn, since scroll-edge effects hide it). Limit: inside the glass sidebar, text and colors come out without their vibrancy (dark text, white dots); layout and content are right. `PREP_SNAPSHOT_TREE=1` also prints the view hierarchy. `screencapture` and ScreenCaptureKit need permission, and `ImageRenderer` leaves AppKit-backed controls blank, so neither is used.
