- 2026-10-07T09:41:25Z edited by claude-code/opus-5.5: requirement

- 2026-10-07T09:44:16Z claude-code/opus-5.5: Done: complete without --commit allowed for code issues (gate validates a given hash only), guide step and needs '[--commit <ref>]', resolveEvidence + evidence_pending in show, drift from the later of confirmed_commit and the entry's last commit, dirty entries cover uncommitted scope changes. Tests: TestCompleteInTheSameCommit, extended TestDriftAgainstConfirmedCommit. This issue itself is the first completed in one commit with its code.
