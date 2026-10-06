First TUI code; the [Stack](/decisions/stack.md) fixes the libraries: Bubble Tea, Bubbles, Lip Gloss (pinned by Glamour v1.0.0 to a lipgloss v1.1.1 pseudo-version), Glamour. The TUI lives in the same binary ([Architecture](/components/architecture.md)) as a human interface next to the CLI, reading through the same domain query engine.

Packages:

- `internal/tui/theme`: tokens and text styles. A `Theme` is built from a `*lipgloss.Renderer`, never from globals, so tests render deterministically (fixed color profile and dark/light background) and the app uses the renderer of its output. Colors are `lipgloss.CompleteAdaptiveColor` (true color, ANSI256, ANSI per light and dark). State colors map `domain.State` values; kind colors map `domain.Kind`.
- `internal/tui/ui`: components as small types or functions taking the theme and plain values (strings, `domain.State`, widths), returning strings. No component reads project data or holds Bubble Tea state; screens in later issues own state and compose components. Markdown goes through Glamour with an `ansi.StyleConfig` derived from the tokens.
- Layout helpers in `ui` split a width or height into panes with minimums, so screens adapt down to 80×24.
- Gallery: `ui.Gallery(theme, width)` renders every component in every variant from fixed sample values. `prep tui --gallery` shows it in a scrollable Bubble Tea viewport that re-renders on resize. `prep tui` without the flag is built by the issue views; until then it explains that.
- CLI: register `tui` as a read command in `internal/cli/cli.go`; it needs a terminal.
- Snapshot tests in `internal/tui/ui`: render the gallery at 80 and 120 columns with a fixed true-color dark renderer and a fixed light renderer, compare with `testdata/*.golden`; `go test ./internal/tui/... -update` rewrites them.
