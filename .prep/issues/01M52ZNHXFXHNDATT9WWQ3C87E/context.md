How this release is worked, as of 2026-10-17. Decisions behind it: the rethink parent 01M4ZCK8C17RDGBVCTMQPGMEK7 and its area issues; the handoff decision 01M52ZR8564ECW8GP3ZHSXEM0A (with its logged findings).

Roles: the main session (Claude Code) defines issues with the user, readies them, hands them off, checks, completes and merges. Codex implements, to save Claude usage (the user's preference); a Claude subagent takes work Codex should not. Implementers log 'Implemented, ready for check' and stop; they never complete, commit or push.

One issue at a time with the user: draft the requirement and criteria, get the user's OK, then define, write a context with a file map, add criteria phrased as checks, ready. Every criterion that names a behavior needs a test.

Handoff: `just install` first (the installed prep must match main), then `sh /Users/azaharias/coding/prep-wt/handoff-codex.sh <issue-id> <short-name>` with run_in_background. It creates worktree prep-wt/<short-name> on branch handoff/<short-name> and writes logs to prep-wt/logs/. Without Codex usage, implement with a Claude subagent (Agent tool, isolation worktree, the same prompt as handoff-codex.sh): Haiku for small, well-mapped issues, Sonnet for subtle code; its worktree is under .claude/worktrees, merge its branch with --no-ff and remove the worktree after. Do not edit handoff-codex.sh while it runs (bash reads it as it goes). At most two implementers at a time: more overload this machine and test suites time out.

Check: read the diff, run go test ./..., prep check (build ./prep from the branch), prep fmt --check, and verify each criterion yourself. Drift and confirm are gone with the knowledge model; descriptive entries are converted by 01M52ZRE24EMGBHX3SKV62436G. Commit in the branch, then merge into main with --no-ff (never rebase: it breaks confirmed_commit). Push, wait for GitHub CI on that commit (check the run's head SHA), run just install so the installed prep matches main, then complete.

Completion: check the criteria, then prep complete <id> --commit <merge hash> for code issues. Decisions made along the way go in with prep decide [<issue>] --title --description --body-file (records are never edited; supersede instead). Decision and research issues complete from defined once their linked record exists.

Records of removed features: no prep command deletes them and the record hook blocks agents. `python3 -I /Users/azaharias/coding/prep-wt/convert-records.py <stage> <root>... [--dry-run]` converts them, run only with the user's go-ahead, after the new binary is installed, in this repository and the other prep projects (prep projects lists them). Stages review, contract, layout and knowledge have run everywhere. In other projects, commit only the files a conversion touches (the user has unrelated work pending there) and never push.

Live eval runs (prep eval run) need the user's OK every time: propose cases, models, N and cost, then wait. Implementers never run them. Writing eval cases needs no OK.

Known: a hook false positive blocks commands whose text contains the records directory path, even reads or prep writes; rephrase. Flaky tests: TestLiveRunner, TestStatuses/timeout and TestEvalMutate under load (logged on 01M52ZQ5YCA96FHQT9WQB6M8C2).

Next: the knowledge conversion of this repository's descriptive entries (01M52ZRE24EMGBHX3SKV62436G, approved tables in its context, AGENTS.md draft there); the guidance issues 01M52ZRM0R2QPKK4HZE6PV62CE, 01M52ZRPZXEJDACVHTHEPX77MY, 01M52ZRSZFBMWZQJPAXVHXD7M1 and 01M52ZRX0B8831K07HHVAQ6S5F (defined, need context and ready; each needs an eval case whose run the user approves); the TUI Knowledge screen 01M52ZRH1Q2ZW3QC5KQ6A12BPC; then eval cases, bound output and the prune sweep.
