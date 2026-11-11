Set the title through Bubble Tea v2, which already emits OSC 2: `tea.View.WindowTitle` (bubbletea v2.0.10, `tea.go` View struct; `cursed_renderer.go` writes `ansi.SetWindowTitle`, and on exit it resets the title to the empty string when one was set, which is why restoring needs separate handling).

Code:
- `internal/tui/app.go` `(*Model).View()` (around line 1154): add `v.WindowTitle = "prep: " + m.projectName()`. `projectName()` in `internal/tui/projects.go` is the name the header shows (falls back to "prep"), so title and header agree. Check whether other screens/views return their own `tea.View` (the gallery in `tui.go` line ~109 is separate; leave it alone).
- Restore on exit: the program cannot read the old title, so use the xterm window-title stack that Ghostty, kitty, iTerm2 and xterm support: write `CSI 22 ; 2 t` (push title; `ansi.PushWindowTitle`-style, else the raw string `"\x1b[22;2t"`) to `out` in `Run` (`app.go` line ~71) before `tea.NewProgram(...).Run()`, and `CSI 23 ; 2 t` (`"\x1b[23;2t"`) after it returns, also on error. Do it only when `out` is a terminal (skip for tests/pipes, e.g. check `out` is an `*os.File` with `term.IsTerminal`, or an Options flag), so piped output has no stray escapes. Terminals without a stack ignore the sequences; then the title just ends up empty or the shell's own prompt title resets it, which is acceptable.
- Switching project (`switchProject` quits and the CLI restarts the TUI in `internal/cli/tui.go`) goes through `Run` again: push/pop pairs balance, and the new Run sets the new project's title. Verify.

Free of decisions: no new key bindings, no config. Not in the knowledge base as a decision; related: [User input: TUI is the overview beside the agent](/decisions/01M529ANGPGWRBEDDD9V03AAE4.md) (several TUIs open beside harnesses, hence the title).

Tests: `internal/tui/app_test.go` or `program_test.go` (they run a real program against a pipe): assert `m.View().WindowTitle == "prep: <project>"`, and that `Run` with a non-terminal writer emits no push/pop escapes; test the push/pop helper separately against a buffer. Checks: `go test ./...`, `just check`.

Size: small.
