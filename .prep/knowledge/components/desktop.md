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
- **Screens**: Issues (saved views menu, filter field taking words and query flags, tree with parents expanded, `/` focuses the filter), detail (state, links to parent, dependencies, children, blocks and knowledge, transitions from `prep guide --json` confirmed in a sheet that takes their flags, criteria as checkboxes, decisions, findings, work log with an add field), Knowledge (entries by type, rendered body, `/…md` links open entries) and Agent (each agent's current issue, the activity feed). ⌘1–3 switch screens, ⌘R reloads, ⌘O opens a project.
- **Projects**: the toolbar dropdown lists the [project registry](/components/projects.md) merged with projects the app opened itself (older binaries have no registry); the window opens the launch directory (`PrepDesktop <dir>`, `open -a Prep <dir>`), else the last project.
- **Build**: `just desktop` assembles `desktop/build/Prep.app` (Info.plist from `desktop/Resources`, ad-hoc signed); `just desktop-test` runs `swift test` against `./prep`; CI runs both on `macos-15`. Developer ID signing, notarization and a Homebrew cask are not done yet.
- **Seeing the app (agents)**: `PrepDesktop <project> --snapshot out.png [--select <id>] [--tab issues|knowledge|agent] [--entry /path.md] [--size WxH] [--appearance light|dark]` (or `just desktop-snapshot out.png …`) loads a real window off-screen and writes a PNG without Screen Recording permission: `cacheDisplay`, plus compositing for Liquid Glass, which the window server draws (flat stand-ins for glass panels, their content drawn on top, list rows redrawn, toolbar items last). Limit: inside the glass sidebar, text and colors come out without their vibrancy (dark text, white dots); layout and content are right. `PREP_SNAPSHOT_TREE=1` also prints the view hierarchy. `screencapture` and ScreenCaptureKit need permission, and `ImageRenderer` leaves AppKit-backed controls blank, so neither is used.
