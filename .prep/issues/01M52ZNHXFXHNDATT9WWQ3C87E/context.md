How this release is worked, as of 2026-10-17. Decisions behind it: the rethink parent 01M4ZCK8C17RDGBVCTMQPGMEK7 and its area issues; the handoff decision 01M52ZR8564ECW8GP3ZHSXEM0A (with its logged findings).

Roles: the main session (Claude Code) defines issues with the user, readies them, hands them off, checks, completes and merges. Codex implements, to save Claude usage (the user's preference); a Claude subagent takes work Codex should not. Implementers log 'Implemented, ready for check' and stop; they never complete, commit or push.

One issue at a time with the user: draft the requirement and criteria, get the user's OK, then define, write a context with a file map, add criteria phrased as checks, ready. Every criterion that names a behavior needs a test.

Handoff: `just install` first (the installed prep must match main), then `sh /Users/azaharias/coding/prep-wt/handoff-codex.sh <issue-id> <short-name>` with run_in_background. It creates worktree prep-wt/<short-name> on branch handoff/<short-name> and writes logs to prep-wt/logs/. At most two implementers at a time: more overload this machine and test suites time out.

Check: read the diff, run go test ./..., prep check (build ./prep from the branch), prep fmt --check, and verify each criterion yourself. Look for stale knowledge with prep knowledge show of likely entries: knowledge find is bounded and misses matches. Confirm drifted entries only after reading them. Commit in the branch, then merge into main with --no-ff (never rebase: it breaks confirmed_commit). Push, wait for GitHub CI on that commit (check the run's head SHA), run just install so the installed prep matches main, then complete.

Completion until the contract removal lands: `python3 -I /Users/azaharias/coding/prep-wt/complete.py <spec.json>` checks criteria and records evidence and knowledge reviews from one spec; issues readied before the policy existed have no bound contract and complete directly with prep complete --docs.

Records of removed features: no prep command deletes them and the record hook blocks agents. `python3 -I /Users/azaharias/coding/prep-wt/convert-records.py <stage> <root>... [--dry-run]` converts them, run only with the user's go-ahead, after the new binary is installed, in this repository and the other prep projects (prep projects lists them). Stage review exists; the contract removal adds a contract stage. Commit nothing in other projects.

Known: a hook false positive blocks commands whose text contains the records directory path, even reads or prep writes; rephrase. Flaky tests: TestLiveRunner, TestStatuses/timeout and TestEvalMutate under load (logged on 01M52ZQ5YCA96FHQT9WQB6M8C2).

Next: Codex implements the contract removal (01M52ZQAGVQ5HX0GTBX7R2P47G) in prep-wt/contract-removal; check it, add the contract stage to convert-records.py from the record files the implementer logs, convert with the user's go-ahead, complete with complete.py. The gates decision (01M52ZR56WW8WDH9QGKJFN32FP) has its outcome; its implementation 01M54KB1FPD39RBHZTAK5NSWVM is drafted and waits for the user's OK. Complete the decision issue itself once that lands (today it still needs ready and claim).
