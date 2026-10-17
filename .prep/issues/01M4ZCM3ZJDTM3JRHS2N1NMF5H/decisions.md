## D1: User input: done is settled in the session
date: 2026-10-16

User, 2026-10-16. Whether an issue is done is discussed in the session between agent and human: the agent asks the human to review, or the human asks for an issue to be closed. Many issues may need no explicit review: when the requirement was clear and the Definition of Done runs the test suite, there may be no acceptance step at the end. To be discussed. Approving in the TUI Inbox disconnects agent and human.

## D2: Remove human approval records and the Inbox
date: 2026-10-16

User decision, 2026-10-16. Done is settled in the session between agent and human; the human needs no approval records in prep and no Inbox. Either the agent signals that it needs the human, or the human asks for work and the agent says what is needed; the issue list shows each state. Remove review approve, resolution reviews, approval prerequisites on dependencies, dependency override, the Inbox screen and its completion bundle (about 1,500 lines across domain, CLI and TUI). This also restores design principle 2: prep has no approval mechanism of its own.

## D3: Keep a minimal reopen
date: 2026-10-16

User agreed with the recommendation, 2026-10-16. Reopen when the requirement still holds but the result is wrong, for example when the checking agent sends a subagent's work back: criteria, context and decisions stay valid. File a follow-up issue when the requirement changes or the work is new. Minimal means: reopen moves the old resolution into the issue's history and returns the issue to ready; no resolution generations, fingerprints or review binding. Reopen was used twice since it was added.

## D4: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16/17, are carried into release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) by 01M52ZQDGE4WPA096TW2EHMZ4N. Open questions raised here were settled by the release's decision issues.
