## Summary
Added a project-specific terminal title to the TUI and balanced terminal title-stack push/pop sequences around its run. Changes are in the current working directory.

## Decisions
### Restrict title-stack escapes to terminal files
Only push/pop the xterm title stack when the output is an `*os.File` attached to a terminal. This avoids escape output for pipes and test writers; the pop is deferred so it also runs when the TUI exits with an error.

## Findings
### Project title source
`Model.projectName()` is the same source used by the header and falls back to `prep`; the view title uses it directly.

## Problems
The handoff requires updating knowledge entries describing changed behavior, but separately prohibits reading or changing `.prep` except for this report. I left the knowledge base untouched to follow that restriction.

## Checks
- `go test ./...` — passed.
- `just check` — passed.
