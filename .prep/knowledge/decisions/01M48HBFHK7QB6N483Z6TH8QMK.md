---
type: decision
title: "Hit-test with Lip Gloss v2 layers and Compositor.Hit"
description: "Hit-test with Lip Gloss v2 layers and Compositor.Hit"
date: "2026-10-06"
issue: 01M4863WN8D1HHTP302Q8G2PQX
supersedes:
  - 01M48HBFHGHMJSWWPDB05N7ZAS
---

After the Charm v2 migration (01M48N7E9GRPWH1SKX5MTB931Y), the screen is composed with Lip Gloss v2 layers. Clickable elements get a layer with an ID that names them: `tab:<n>`, `row:<issue id>`, `rel:<n>`, `pane:list`, `pane:detail`, `know:<path>`, `set:<n>`, `menu:<n>`, `pick:<n>`, `crit:<n>`, `kind:<n>`, `field:<n>`, and `dialog` for the dialog as a whole. Each layer sits at the position where it is drawn. The open dialog is a layer with a higher Z, so `Compositor.Hit(x, y)` returns its elements first, and a hit on `dialog` without a more specific element does nothing. The model keeps the compositor from the last `View` and resolves `tea.MouseClickMsg` and `tea.MouseWheelMsg` with `Hit`, then calls the same functions the keys call.

This is Charm's own mechanism. It replaces the plan to record geometry by hand (D1) and the string splicing in `ui.Overlay`. Alternatives considered: bubblezone, a third-party library whose markers break under the overlay, and hand-recorded geometry, which would duplicate what layers already provide.
