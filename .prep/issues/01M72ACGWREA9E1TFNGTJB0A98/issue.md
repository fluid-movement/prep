---
title: Before-tool hook names the decisions scoped to a file when the agent first edits it
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
tags:
  - hooks
priority: high
---

When an agent is about to edit a code file, the before-tool hook tells it which current decisions are scoped to that file: title and record ID, at most five, once per session for each decision. The edit is not blocked.

Why: the rule "before changing code, look for decisions about it" depends on the agent's judgment and is mostly skipped. A knowledge lookup preceded the work in 0% (daypack), 12% (fpa-events) and 29% (prep) of the turns that changed code (finding 01M72A9TBJ5JQXT2VGT15N2HW1). Prep already knows which decisions apply, because decisions carry scope globs, but it says so only in the Stop question, after the files have been changed (30 of 81 Stop questions listed scoped decisions). Moving that list to the moment of the first edit makes the lookup happen every time, costs about 30 tokens per decision and adds no request.

The Stop question no longer lists scoped decisions, since the agent has seen them by then. The session rule about looking up decisions stays, because decisions without a scope are project-wide and are not surfaced by this hook.

The hook covers the file-editing tools of the harness. Edits made through shell commands or by a separate process are out of scope.

## Open questions
