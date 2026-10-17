---
type: finding
title: "Findings: Ledger analysis: is prep used in Claude Code sessions when we expect it"
description: "Findings: Ledger analysis: is prep used in Claude Code sessions when we expect it"
date: "2026-10-10"
issue: 01M4J8780KZRKWHJR4NN8AD85M
---

## Is the ledger being written?

Yes, but it holds little history. The token-ledger plugin's bridge now lives in the prep plugin (`prep@prep`, `hooks/register.ts`) and records every tool call, request and prep command. The stream starts 2026-10-08 16:23 UTC: an agent deleted it during a test (`rm -f .prep/local/activity.jsonl`, session 832549f2), and it is capped at 1 MiB, pruned to its newest 2,000 lines, which is two to three days at this repository's rate. This analysis therefore used Claude Code's transcripts for the full history (13 main sessions plus subagents, 2026-10-05 to 2026-10-10) and the stream for lifecycle sequences since 2026-10-08. 34 prep events have no session (prep run outside the bridge's environment), and prep events carry the actor `claude-code/opus-5.5` while bridge events carry `claude-code`, so they join by session only. → 01M4J8JVQ7TVGMEF0BB2AZRFG8 (retention), a precondition for measuring hooks.

## Misses by kind

| Kind | Before the command existed | Avoidable (command existed) | Cause | Change |
| --- | --- | --- | --- | --- |
| Knowledge reads by file (`cat`, `grep`, `sed` on `.prep/knowledge`) | 108 | 5 | Before 2026-10-08 there was no other way. Since then, only frontmatter checks (`confirmed_commit`, `scope`) and one diff: `knowledge list` doesn't show the confirmation state | 01M4J8JW0Z4746FVC99577A6QX; redirect hook in 01M4J8789E81QAJHG8R8ZPEMRD |
| Knowledge writes through scripts | 6 | 25 before 2026-10-08, 11 since (6 scripted, 5 full-body heredocs) | Missing capability: `knowledge update` replaces the whole body, so sessions reinvent `kedit.py` (raw file read + old/new replacements + full body piped back), at least five times | 01M4J8JVK1GW7121JQAC974X1J |
| Issue record reads by file (`cat issue.md`, `acceptance.md`, `decisions.md`, `ls .prep/issues`) | — | 67 (all but ~5 before 2026-10-08) | `prep guide`'s `Read:` section names the record files, so the guide itself sends agents to them | 01M4J8JVW74HVE79K5QDTV930Z; redirect hook |
| Issue record writes by hand | — | 9 | All on 2026-10-05/06 while the write commands were being built; none since | none |
| Lifecycle without `prep guide` | — | 21 of 24 completions, 19 of 20 claims in the stream | Agents file the issue themselves from the conversation, then run new → define → context → criterion → ready → claim → criterion → complete within one or two minutes; they wrote the requirement, so the guide looks redundant, but its knowledge candidates and gates go unseen. Some issues were filed after the fix (audit fixes, 2026-10-08 17:06–17:46) | open question on 01M4J8789E81QAJHG8R8ZPEMRD |
| Skill not loaded | — | 1 of 12 working sessions | The skill loads almost only when the person types `/prep`; the session without it (91ba8392) reinvented `kedit.py` and read records by file | skill-loading hook in 01M4J8789E81QAJHG8R8ZPEMRD |

Subagents did well once the briefing told the parent to pass the rules on: on 2026-10-08 they made 14 `prep knowledge` reads and no knowledge-file reads; their only `.prep` reads were `.prep/.gitignore`, which has no prep command.

## What this says for 0.3.0

The briefing and the knowledge read commands already did most of the work: knowledge-file reads fell from 108 to 5 once `find|show` existed and the briefing named them. What is left splits in two. Missing capabilities (targeted knowledge edits, confirmation state in `knowledge list`) and the guide's own pointer to files cause misses that no hook can fix, so they get code issues. The behavioural misses (stray reads, the skill not loaded, the guide skipped) are what the Claude Code hooks are for, and measuring them before and after needs the stream to keep more than three days.

Method: a script over the transcripts classified every Read/Grep/Glob/Edit/Write on `.prep` and every shell segment touching `.prep` (heredoc bodies, git, prep, echo/printf and test fixture paths excluded), deduplicated by tool-use ID, and compared each timestamp with the commit that introduced the matching prep command; the stream gave each issue's op sequence per session.
