## D1: User input: TUI is the overview beside the agent
date: 2026-10-16

User, 2026-10-16. The TUI is used for tracking the issue list, the state of the project and what the agent is working on, open beside the harness. Issues were created in the TUI, but agent-written issues are never edited there; simple editing stays anyway. The current Inbox makes no sense: code changes are not reviewed in the TUI (better tools exist), and approving there disconnects the human from the agent.

## D2: TUI only: issue screens, Agent screen, decision lookup; desktop and extra themes go
date: 2026-10-16

User decisions, 2026-10-16. TUI only: delete the desktop app (about 6,600 lines of Swift) now that the Inbox it was to hand over is removed; its open issues (01M4JVP89F3PGDHYETHEPS3YBB, 01M4JVPRARK7X2MB05BNSD8WAA, 01M4JVPRRGXSKRB2AY8PJSHSP3, 01M4JX2N231RDG9KEHDBC1ZRE1) are dropped and 01M4TSY329F505HDC22HFYPD42 becomes the deletion, when actionable work is created. Keep the issue screens (list, views, filters, tree, detail), simple editing, mouse and the Agent screen including its Usage tab (useful, cheap, in its own tab). Keep the Knowledge screen as a read-only lookup of decision and finding records, showing the links from an issue to the decisions made in it. Themes: keep one dark and one light theme; remove the other built-in themes, custom themes, prep theme new and the theme preview. Remove the Inbox. Fix the live refresh bug (01M4WYKSGGDM7R3FSKE8451X2X).

## D3: Activity stream is the Agent screen's live feed only
date: 2026-10-16

Agreed with the area, 2026-10-16. Evals measure agent behavior, so the stream needs no long history: keep the current retention and drop 01M4J8JVQ7TVGMEF0BB2AZRFG8 when actionable work is created. prep focus stays (sets the agent's current issue). prep watch stays only if something outside the TUI still uses it.

## D4: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16/17, are carried into release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) by 01M52ZQ7J4SGWXY2W9X8EPEJV1, 01M52ZQPFMW4CHE24H7BFWHSTD, 01M52ZRH1Q2ZW3QC5KQ6A12BPC. Open questions raised here were settled by the release's decision issues.
