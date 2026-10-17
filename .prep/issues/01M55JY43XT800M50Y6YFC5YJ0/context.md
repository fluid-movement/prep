From eval run 20261017-183812-135e (see the finding linked to this issue).

Where things are:
- Rules: `internal/cli/rules.md` (one rule per line; `internal/cli/skill.md` ends with the same text, a test keeps them equal; run `just sync-skill` after). Change the decision rule to: 'Record a decision with prep decide when it is made, including one the user states, with the reason given; your own small implementation choices need no record.' Add a rule before it: 'Before changing code, look for decisions about it with prep knowledge find <area>; follow them, or record a new decision with prep decide --supersedes <id> first.'
- Finish hook wording in `internal/cli/hook_finish.go`: mention the user's stated choices in the question ('did this turn make or receive a decision ... including one the user stated').
- Delete the obsolete cases `eval/cases/implement/complete.yaml` and `eval/cases/implement/log-and-stop.yaml` (covered by checker-completes and handoff-from-guide).
- `eval/cases/natural/fix-matching-issue.yaml`: replace the rubric item that expects the descriptive entry about profile fields to be edited with one that expects the changed limit to be recorded as a decision (prep decide) or explained; keep the other expectations.
- Tests: the rules and skill tests keep passing; a finish hook test checks the new wording.

Checks before stopping: `go test ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`, `./prep eval list` loads the cases. Do not run `prep eval run`; the checker proposes a rerun of record-code-decision, respect-decision and fix-matching-issue to the user.
