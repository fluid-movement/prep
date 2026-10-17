Removal, not redesign: delete code and tests that only serve the removed themes; keep behavior for default and light unchanged.

Where things are:
- `internal/palette/palette.go`: `Builtins`, `builtin`, `Names`, `IsBuiltin`, `Resolve`/`resolve` (custom themes and `base` inheritance), `Validate`, `Def`, `Pick` (automatic dark/light by terminal). Keep two palettes and `Pick`; drop custom-theme resolution and `Def` if nothing else needs them.
- `domain.ThemeDef` in `internal/domain/types.go` and the `themes` config key parsed in `internal/mdstore/store.go`: remove; an unknown `themes` key then fails config validation like any unknown key (P003).
- `internal/cli/theme.go` and its entry in the command table in `internal/cli/cli.go`: remove `prep theme`.
- TUI settings: the theme preview (`themePreview`, `m.preview`, `previewKey`, `previewKeys` and the `bind("Settings", ...)` text about previewing themes in `internal/tui/app.go`, plus related code in `internal/tui/tui.go`, `nav.go`, `editing.go`): the theme setting toggles between the two themes directly.
- Goldens: delete `internal/tui/testdata/theme-preview-110x28.golden` and the gallery goldens for the six removed themes in `internal/tui/ui/testdata/`; regenerate changed goldens with the tests' update flag (see how the tests define it) and review the diff.
- `prep tui --gallery` keeps working with t/T switching between the two themes.
- Docs: `README.md`, the skill (`internal/cli/skill.md`, copied to `plugins/claude-code/skills/prep/SKILL.md` and `plugins/codex/skills/prep/SKILL.md`), and knowledge entries found with `prep knowledge find theme` and `prep knowledge find palette` (at least [Palette and themes](/components/palette.md), [TUI design system](/components/tui-design-system.md), [Storage format](/conventions/storage-format.md), [TUI editing and keys](/components/tui-editing.md)). Update knowledge only through `prep knowledge update`. If this repository's committed config template has a themes key, say so in the log; never hand-edit records.

Checks before stopping: `go test ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`, `git grep -i -E 'catppuccin|gruvbox|nord|pastel|monochrome|high-contrast'` finds nothing outside issue records.
