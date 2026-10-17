How this release is worked, as of 2026-10-17. Decisions behind it: the rethink parent 01M4ZCK8C17RDGBVCTMQPGMEK7 and its area issues; the handoff decision 01M52ZR8564ECW8GP3ZHSXEM0A (with its logged findings).

Roles: the main session (Claude Code) defines issues with the user, readies them, hands them off, checks, completes and merges. Codex implements, to save Claude usage (the user's preference); a Claude subagent takes work Codex should not. Implementers log 'Implemented, ready for check' and stop; they never complete, commit or push.

One issue at a time with the user: draft the requirement and criteria, get the user's OK, then define, write a context with a file map, add criteria phrased as checks, ready. Every criterion that names a behavior needs a test.

Handoff: `just install` first (the installed prep must match main), then `sh /Users/azaharias/coding/prep-wt/handoff-codex.sh <issue-id> <short-name>` with run_in_background. It creates worktree prep-wt/<short-name> on branch handoff/<short-name> and writes logs to prep-wt/logs/. Without Codex usage, implement with a Claude subagent (Agent tool, isolation worktree, the same prompt as handoff-codex.sh): Haiku for small, well-mapped issues, Sonnet for subtle code; its worktree is under .claude/worktrees, merge its branch with --no-ff and remove the worktree after. Do not edit handoff-codex.sh while it runs (bash reads it as it goes). Confirm drifted knowledge after a merge commit, not during conflict resolution, or the stamps point before the merge. At most two implementers at a time: more overload this machine and test suites time out.

Check: read the diff, run go test ./..., prep check (build ./prep from the branch), prep fmt --check, and verify each criterion yourself. Look for stale knowledge with prep knowledge show of likely entries: knowledge find is bounded and misses matches. Confirm drifted entries only after reading them. Commit in the branch, then merge into main with --no-ff (never rebase: it breaks confirmed_commit). Push, wait for GitHub CI on that commit (check the run's head SHA), run just install so the installed prep matches main, then complete.

Completion: check the criteria, then prep complete <id> --docs <entry>... (or --no-impact <reason>), with --commit <merge hash> for code issues. The contract machinery and complete.py are gone.

Records of removed features: no prep command deletes them and the record hook blocks agents. `python3 -I /Users/azaharias/coding/prep-wt/convert-records.py <stage> <root>... [--dry-run]` converts them, run only with the user's go-ahead, after the new binary is installed, in this repository and the other prep projects (prep projects lists them). Stages review and contract have run everywhere. In other projects, commit only the files a conversion touches (the user has unrelated work pending there) and never push.

Live eval runs (prep eval run) need the user's OK every time: propose cases, models, N and cost, then wait. Implementers never run them. Writing eval cases needs no OK.

Known: a hook false positive blocks commands whose text contains the records directory path, even reads or prep writes; rephrase. Flaky tests: TestLiveRunner, TestStatuses/timeout and TestEvalMutate under load (logged on 01M52ZQ5YCA96FHQT9WQB6M8C2).

Next: a Sonnet subagent implements the lock (01M52ZQKG3PZZ6YS5P4Z81FKGC); check it. Then the import/migrate removal (01M52ZQSEP2EKTYKYFC79EGFS6, ready, waits on the lock), then enrich the record layout (01M52ZQZBAS48ZK64MW7H85QJ4). The eval harness errors fix is merged; complete it once CI on main is green. Codex usage resets at 15:08 on 2026-10-17.
