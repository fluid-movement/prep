- 2026-10-13T09:21:59Z edited by codex: parent

- 2026-10-16T18:33:33Z edited by claude-code: parent

- 2026-10-16T18:39:53Z edited by claude-code: depends_on, requirement

- 2026-10-17T18:19:21Z codex: Measurements (Unicode characters, installed prep baseline -> new source binary, same current records): knowledge list 56,577 -> 5,871; list 32,869 -> 1,055; knowledge find lifecycle 3,064 -> 2,762; knowledge find knowledge 6,652 -> 3,000; show release parent 01M52ZNHXFXHNDATT9WWQ3C87E 7,694 -> 5,097; help 4,222 -> 4,206 (build version string; help code unchanged); prime 2,511 -> 2,511; guide this issue 6,773 -> 6,773; check 56 -> 56. Initial measurements before adding the updated knowledge decision: knowledge list 56,293; list 32,869; find lifecycle 3,064; find knowledge 6,404; show parent 7,694; help 4,222; prime 2,511; guide 6,680; check 56. Context baseline was knowledge list 56,472, list 32,869, find 7,400-7,800 (query unspecified), show parent 7,683, help 4,224, prime 2,765, guide 2,405, check 56. Updated decision 01M55HBDFHYGAH7DGSJNWXVEW1 supersedes the earlier output-bounds decision. TestShippedCases loads and compiles the scale/session cases and assertions; later approved live eval should cover scale-find-issue, scale-what-next and session-what-next. No live eval run. Sandbox git auto-staging emitted index.lock permission warnings, but record writes succeeded; no commit or push attempted.

- 2026-10-17T18:19:36Z codex: Validation passed: focused output-bound tests; go test ./internal/cli ./internal/domain; go test ./... (including TestShippedCases and internal/update; the known sandbox localhost-listener failure did not occur); go build -o prep ./cmd/prep; ./prep check (248 issues, 306 knowledge entries, 0 errors, 0 warnings); ./prep fmt --check; git diff --check. GitHub CI on main remains for the checker; no live evals, criterion checks, completion, commits or pushes were run.

- 2026-10-17T18:19:37Z codex: Implemented, ready for check: bounded default issue and knowledge text reads with recovery footers, preserved complete JSON lists, and added bound/recovery tests plus updated knowledge and measurements.
