---
type: finding
title: 'Latest sessions: prep work is 9-14% of cost; judgment steps are skipped, mechanical ones hold'
description: 'Measured 2026-11-10 from transcripts (daypack 3 sessions + 39 subagents, fpa-events since Oct 31, prep since Oct 17). Updates the 2026-11-07 finding: prep-only turns fell to 2-7% of cost because agents batch in shell. Prompt text cannot identify the issue (under 10% hit rate); Stop asks pay off only when 3+ files changed; show follows guide only when guide omits the requirement.'
date: "2026-11-10"
issue: 01M5FS7KX63E8KZARSXXAS44C4
---

Method: Claude Code transcripts including subagents, requests deduplicated by message id, list-price weights (cache read 0.1, cache write 1.25, output 5), about 3.5 characters per token, about ±30%. This updates finding 01M6V6R1BBWFSACRAAZX1GPAGP, whose 18-22% figure for prep-only turns no longer holds for sessions after 0.3.0.

## Cost

| share of session cost | daypack | fpa-events since Nov 6 |
|---|---|---|
| hook text including re-reads | 0.5% | about 1% |
| requests that only run prep commands | 6.7% | 2.3% |
| Stop question (one full request each) | 1.8% | about 3% |
| hook denials and retries | 0.9% | 0.7% |
| all prep work including calls mixed into other requests | about 12-14% | about 9% |

Agents batch on their own now: in daypack 51 Bash commands wrap prep in loops or helper functions and 50 requests carry three or more prep calls. Real command failures are 1-3% and nearly all are hook denials, not gates. A run of about 8 denials in one session (echo hi, prep check, git status) is bug 01M5A68QTP1ZHEEXW9NY8BEBYB. Failed commands already record their error code in the activity stream.

## What holds and what does not

Mechanical steps hold: 38 of 39 subagents start with prep guide because the hand-off prompt contains the command; record redirects work.

Judgment steps are skipped. In turns that edited code or delegated work: issue lookup before the work 19% (daypack), 34% (fpa-events), 53% (prep); knowledge lookup before the work 0%, 12%, 29%; no prep command at all 22%, 41%, 12%. These are per turn, so multi-turn issues make them look worse than they are. Clear untracked in-repo fixes are few (2 in fpa-events); in both the agent searched, found nothing and went on without creating an issue.

The prompt hook sends its text once per session (6 sends for 80 prompts in daypack), so it acts as a second session-start rule.

## Tested and rejected

Prefilling the prompt hook with issues matched from the prompt text: replayed every user prompt against the project's issue titles (rare-word weighted, top 3). The issue the agent then opened was in the top 3 for 4 of 25 (daypack), 8 of 72 (fpa-events) and 2 of 47 (prep) first touches, while 75-82% of prompts got some match. Prompts are conversational ("sounds good", "move hooks out of v1", "keep going"); the issue follows from the conversation, not the prompt. A classifier that sees only the prompt, whether rules, Jev or a local model, has the same limit; it would need the conversation tail.

Deriving claim and completion from commits: completion depends on checked criteria, which is a judgment; claim is already optional (6 claims for 30 completions in daypack). Nothing left to derive.

A batch apply command: agents already batch with shell, and the content (titles, bodies, criteria) must be written either way.

## Stop question

81 asks across the three projects, 29 followed by prep decide or prep finding (36%). Yield by changed code files: 1-2 files 3 of 26 (12%), 3-8 files 10 of 21 (48%), more than 8 files 16 of 34 (47%). No difference by whether scoped decisions were listed (37% against 35%), whether a person or a task notification started the turn, or whether a decision was already recorded in the turn. Of daypack's 30 decisions, 5 came from the Stop question; the rest were recorded in planning turns under the session rule.

## Guide and show

After a guide that carries the requirement (code issues that are ready or in progress) show followed 3 of 34 times. After a guide without it (other kinds, open, defined, done, stale, parents) show followed 48 of 59 times.

## Other

prep knowledge show <bare record id> fails with E_NOT_FOUND (3 times in these sessions) although decide, finding and find print bare ids. Orchestrators run prep show only to read state (70 shows without a guide in main sessions, usually cut with head or grep).
