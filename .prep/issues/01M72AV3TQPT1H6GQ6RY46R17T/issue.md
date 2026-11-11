---
title: Activity stream records hook failures, agent ids and injected sizes, and readers can include archives
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
tags:
  - usage
priority: high
---

The activity stream gains what later analysis needs and cannot reconstruct:

- A hook that fails open records a hook event with the verb "error" and a short reason. Today every hook error returns silently, so a broken hook looks the same as a quiet one.
- Tool, hook and prep events carry the harness agent id when the event comes from a subagent, so subagent activity can be told from the main session's.
- A hook event that adds text to the model's context records its size in characters.
- `prep activity` can read the archived streams as well as the current one (a flag such as `--all` or `--since <date>`). Today readers see only the current file: on 2026-11-10 the prep stream began that day and daypack's on 2026-11-09, though both projects had earlier sessions.

Why: the usage research (01M5FS7KX63E8KZARSXXAS44C4) had to read harness transcripts because the stream lacked these, and the activity report (01M72ADYEGWG0QN92F653MD24F) can only be as good as the stream.

Not in scope: request events. The stream held 125 request events for daypack against several hundred requests in the transcripts for the same days; why is a question for the report issue to investigate first.

## Open questions
