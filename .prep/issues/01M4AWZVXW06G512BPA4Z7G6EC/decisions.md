## D1: State colors move to the palette package
date: 2026-10-08

The CLI must print the colors the TUI derives without importing the TUI (and its Bubble Tea dependencies) into the read path, and the plugin must never repeat the mapping. Moving `stateTokens` and the kind and border tokens to `internal/palette` gives one source both use. Alternative considered: compute in the CLI from a copy of the map, rejected since it is the duplication the requirement forbids.
