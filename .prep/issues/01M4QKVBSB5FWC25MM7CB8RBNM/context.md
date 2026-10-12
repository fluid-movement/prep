New package `internal/eval/evalcase` (named to avoid the `testing` word), YAML via `gopkg.in/yaml.v3` like the fixture spec ([Test fixtures](01M4QKVB5EW2N6Y9BDVCFJDC7Q)).

- **Shape**: `Case{Name, Area, Description, Prompt, Fixture (inline spec | ref), Workspace (inline | ref), Models, Variants, Tags, Expect{Tools []ToolAssert, Outcome []OutcomeAssert, Rubric []string}, Limits{MaxTurns, MaxBudgetUSD, Timeout}}`. Assertion kinds are defined by [Tool-use assertions](01M4QKVC39NZBH95845QPC1D5Q) and [Outcome assertions](01M4QKVC87EPF7DT9AN3NTGF6B); this package decodes each into the typed value those packages register, so adding a kind does not change the loader.
- **Fixtures by reference**: `eval/fixtures/<name>.yaml` shared across cases, so twenty cases do not repeat the same project.
- **Layout**: `eval/cases/<area>/<name>.yaml`; `Load(dir)` walks it; `Select(cases, filter)` filters by area, name glob and tag.
- **Validation**: expand aliases against the fixture while loading (using `fixture.Spec` alias list, no build needed), yaml.v3 `KnownFields(true)` for unknown keys, line numbers from yaml nodes.
- **Tests**: valid and invalid case files in `testdata/`, each invalid one naming the error.
