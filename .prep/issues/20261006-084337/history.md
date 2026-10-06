- 2026-10-06T12:58:10Z edited by claude-code/2.1.291: requirement

- 2026-10-06T13:08:13Z edited by claude-code/2.1.291: depends_on

- 2026-10-06T13:24:30Z edited by claude-code/2.1.291: requirement

- 2026-10-06T13:37:08Z claude-code/2.1.291: Hit-testing as in D4, with one simplification: the layers are used for hit-testing only (blank layers with IDs recorded while rendering, kept in m.hits); the screen itself is still rendered as strings, so no golden changed apart from the settings row. Options got NoMouse rather than Mouse (D5), so the zero value keeps capture on.

- 2026-10-06T13:37:08Z claude-code/2.1.291: Tests: mouse_test.go drives clicks and the wheel through Update (tabs, rows, relations and pane focus at 110x28 and 80x24; wheel; knowledge and settings rows; every dialog kind; mouse off); program_test.go runs the real Bubble Tea program on SGR escape sequences. A mutation (row marks one line low) fails 3 tests. go test -race ./internal/tui/... passes.

- 2026-10-06T13:37:08Z claude-code/2.1.291: Menus list only available actions (openMenu filters), so the unavailable-entry reason is tested on a constructed menu; runAction guards it for clicks as for keys. Clicks are ignored while the filter bar is open.
