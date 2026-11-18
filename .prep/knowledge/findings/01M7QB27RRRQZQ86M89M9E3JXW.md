---
type: finding
title: Free opencode models on a handed-off small issue
description: Four free models matched GPT on tests and message text; three missed the write-nothing behaviour that GPT caught.
date: "2026-11-18"
---

# Free opencode models on a handed-off small issue

Trial 2026-11-18: four free opencode models implemented the same issue (01M7Q2AQYM5W9XZXK9BJ43SZYM, a newer-schema message plus a hook line plus "other commands write nothing") from the same commit (f124924, the issue ready and unimplemented), each in its own worktree with the identical prompt that the GPT run (`openai/gpt-6.1-sol`) got. Judged by running each built binary against a project with schema 2, plus the usual checks. One task, one run per model: a first signal, not a ranking.

| Model | go test / vet / prep check / fmt | Message names update, pin, cloud script | Hook line | Writes nothing on rejection | Logged "ready for check" |
| --- | --- | --- | --- | --- | --- |
| gpt-6.1-sol (reference) | pass | yes | JSON additionalContext | yes | yes |
| big-pickle | pass | yes | JSON additionalContext | **no** | yes |
| deepseek-v4.1-flash-free | pass | yes | JSON additionalContext | **no** | yes |
| longcat-2.5-free | pass | yes | JSON additionalContext | yes | **no** |
| hy4-flash-free | pass | yes | plain stdout text | **no** | yes |

## What it shows

- All four passed the test suite, vet, `prep check` and `prep fmt --check`, wrote a test for the hook and produced the required message text. None committed, pushed or hand-edited records. On the visible part of the requirement the free models matched GPT.
- The part that separates them is the criterion nobody's tests forced: "other commands write nothing". A rejected command still appended failure events to the local activity stream; GPT's implementer found that and fixed it, and so did longcat. The other three left it, and their own tests passed anyway. The checker has to run the behaviour (a project with a newer schema, compare files before and after) rather than trust the diff and the tests.
- hy4-flash-free printed the hook text as plain stdout where the others and the existing hooks use the structured `additionalContext` output, a deviation a reviewer should catch.
- longcat did the right thing but skipped the final `prep log` line, which a one-line prompt reminder would fix.

## Use

Free models are usable for small issues whose criteria can be checked by behaviour, with a review that actually exercises the criteria; the review model can be a stronger one. They are a fallback for when paid quota is used up, not a first choice for subtle behaviour. hy4-flash-free is reported as unlimited, which makes it the natural default for routine small issues. Repeat the trial on more issues before relying on any ranking.
