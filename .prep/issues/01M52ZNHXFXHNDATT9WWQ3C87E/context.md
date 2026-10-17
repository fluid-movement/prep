How this release is worked, as of 2026-10-17. Decisions behind it: the rethink parent 01M4ZCK8C17RDGBVCTMQPGMEK7 and its area issues; the handoff decision 01M52ZR8564ECW8GP3ZHSXEM0A (with its logged findings).

Roles: the main session (Claude Code) defines issues with the user, readies them, hands them off, checks, completes and merges. Codex implements, to save Claude usage (the user's preference); a Claude subagent takes work Codex should not. Implementers log 'Implemented, ready for check' and stop; they never complete, commit or push.

One issue at a time with the user: draft the requirement and criteria, get the user's OK, then define, write a context with a file map, add criteria phrased as checks, ready. Every criterion that names a behavior needs a test.

Handoff: `just install` first (the installed prep must match main), then `sh /Users/azaharias/coding/prep-wt/handoff-codex.sh <issue-id> <short-name>` with run_in_background. It creates worktree prep-wt/<short-name> on branch handoff/<short-name> and writes logs to prep-wt/logs/. At most two implementers at a time: more overload this machine and test suites time out.

Check: read the diff, run go test ./..., prep check (build ./prep from the branch), prep fmt --check, and verify each criterion yourself. Look for stale knowledge with prep knowledge show of likely entries: knowledge find is bounded and misses matches. Confirm drifted entries only after reading them. Commit in the branch, then merge into main with --no-ff (never rebase: it breaks confirmed_commit). Push, wait for GitHub CI on that commit (check the run's head SHA), run just install so the installed prep matches main, then complete.

Completion: check the criteria, then prep complete <id> --docs <entry>... (or --no-impact <reason>), with --commit <merge hash> for code issues. The contract machinery and complete.py are gone.

Records of removed features: no prep command deletes them and the record hook blocks agents. `python3 -I /Users/azaharias/coding/prep-wt/convert-records.py <stage> <root>... [--dry-run]` converts them, run only with the user's go-ahead, after the new binary is installed, in this repository and the other prep projects (prep projects lists them). Stages review and contract have run everywhere. In other projects, commit only the files a conversion touches (the user has unrelated work pending there) and never push.

Known: a hook false positive blocks commands whose text contains the records directory path, even reads or prep writes; rephrase. Flaky tests: TestLiveRunner, TestStatuses/timeout and TestEvalMutate under load (logged on 01M52ZQ5YCA96FHQT9WQB6M8C2).

Next: Codex implements the Agent view session lifecycle (01M4E9JJBJJVMC66M38936AB6N, prep-wt/agent-sessions) and the gates change (01M54KB1FPD39RBHZTAK5NSWVM, prep-wt/gates); check both. Once the gates change lands, close the gates and handoff decision issues and the rethink areas (01M4ZCK8C17RDGBVCTMQPGMEK7's children) from defined: one outcome decision each, pointing at their release children, then complete; the rethink parent closes with them. Five drafts with no open questions wait for the user's OK to define: links removal, one lock, import/migrate removal with define --trivial, record layout, eval harness errors.
