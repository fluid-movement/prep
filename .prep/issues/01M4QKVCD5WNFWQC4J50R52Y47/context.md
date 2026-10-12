New package `internal/eval/judge`; inputs from [Transcript](01M4QKVBMEVANE3EBQM2V2VKR8) and the outcome diff ([Outcome assertions](01M4QKVC87EPF7DT9AN3NTGF6B)).

- **Call**: `claude -p` with `--model <judge model>` (default sonnet; configurable), `--output-format json`, `--tools ""` (no tools), `--no-session-persistence`, run from a temp dir inside the isolated environment ([isolation](01M4QKVBFKQKCQP6N42BG2K8Z4)), so no hooks or memory affect it. One call per rubric item. The prompt template lives in `eval/judge/prompt.md`; the response is JSON `{"verdict":"pass|fail","reason":"...","evidence":"..."}`, validated, retried once on invalid JSON, else the item is `error`.
- **Input size**: the transcript is condensed (tool calls and clipped results) to fit; the full diff is included.
- **Calibration**: `eval/judge/calibration/*.yaml` holds clipped transcripts with a human verdict per item; `prep eval judge-check` runs the judge on them N times and reports agreement; below a threshold the judge result is marked untrusted in reports.
- **Tests**: a fake `claude` returning canned JSON for parsing, retry and error paths; calibration runs live only with `PREP_EVAL_LIVE=1`.
