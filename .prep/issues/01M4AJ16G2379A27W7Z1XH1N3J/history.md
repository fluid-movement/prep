- 2026-10-07T06:51:50Z edited by claude-code/opus-5.5: requirement

- 2026-10-07T07:43:38Z claude-code/opus-5.5: Done: complete.go (pure completion over QueryFlags at the cursor), picker under the filter bar with ui.Suggestion (also in the gallery), up/down/tab/click, domain.FlagKind replaces the hard-coded value-taking list in parseFilterText (fixes --priority as text search). Tests: TestCompleteFilterWords, TestFilterBarSuggestsValues, filter-picker snapshots at 110x28 and 80x24. Screen snapshots unchanged; gallery snapshots gained the Suggestions section.
