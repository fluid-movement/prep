## D1: X007 is a warning; --project takes link names
date: 2026-10-11

The requirement named K003 for a missing entry. A cross-project link gets its own code (X007) and, like X005, is a warning: the other checkout may be behind. To make the read command in show and guide always work, `--project` resolves the home project's link names before registry names, since the local name is what references use and the target may call itself something else.
