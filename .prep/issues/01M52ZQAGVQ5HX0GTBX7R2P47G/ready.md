---
contract:
  id: 84f5ebac8a6b145b2f7e11bb7ea6c8239e2a84dee1c890fc7be3ac03c8966439
  verification: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
  documentation: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
  ancestors:
    - 01M52ZNHXFXHNDATT9WWQ3C87E
  requirements:
    - id: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
      kind: verification
      text: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
    - id: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
      kind: documentation
      text: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
    - id: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
      kind: dod
      text: go test ./... passes
    - id: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
      kind: dod
      text: prep check reports no errors
    - id: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
      kind: dod
      text: prep fmt --check passes
    - id: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
      kind: dod
      text: knowledge entries describing changed behavior are updated
    - id: a803295794fe3b038064d7936c354860cc7cff4991958540178d8f0978753989
      kind: acceptance
      text: prep help lists no policy, evidence, exception, dod approve or knowledge review command, and git grep finds none of the removed types, gates, I029 or internal/worktree outside issue records
    - id: a9217c6c0cf0b011c2828a39e17070a025188433cff514d1bcddc096ce4e6d2f
      kind: acceptance
      text: Completing an issue requires only state, staleness, checked criteria, the kind's result and --docs or --no-impact; a test completes a ready-to-complete code issue with no contract, evidence or review records
    - id: 36975dd76d9ca62bb5b25bd1b40c016400158242a54a49dea5854a4cd0df24b0
      kind: acceptance
      text: prep dod --add and --remove edit the project Definition of Done in project.md, covered by a test
    - id: f3f5a9759625d58d24e2b55d6675c3bb986882e3eb87b413904ef1488dabec39
      kind: acceptance
      text: The session briefing and prep guide at the check step show the project Definition of Done, covered by tests
    - id: 67814ea641f163b7915e5943888b063d80eadccdd4df0bf2752ac566e5467a94
      kind: acceptance
      text: Knowledge describing the removed parts is deleted or updated, this repository's records are converted, prep check reports no errors, and GitHub CI on main is green
by: claude-code
at: 2026-10-17T09:35:10Z
baseline: 20261017-093353
---
