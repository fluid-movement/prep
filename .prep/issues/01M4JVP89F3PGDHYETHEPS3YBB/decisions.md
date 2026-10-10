## D1: Swift app over the CLI contract
date: 2026-10-10

Mac-only removes the reason for Electron, Tauri or GPUI. None of them can share the Go domain in-process, so they all cross a process boundary anyway; the CLI's `--json` and `prep watch` already are that boundary. Wails (Go plus web UI, domain in-process) was the real alternative and is rejected unless the JSON contract turns out too thin. Reading `.prep/` from Swift is rejected: two domain implementations drift.

Accepted downsides: no good native markdown editor or renderer (use a library or a small WebView for rendering); SwiftUI needs AppKit for the letter-sequence shortcuts and dense lists; agents verify native UI harder than the TUI or a web page (own issue); a second language with a macOS CI runner, signing and notarization.
