---
contract:
  id: ff47c17cfe76e0ce1ea8ad2763498eb653bc91e28cee4663e586bf5a18534178
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
    - id: 96928a59567d66910c32b3754e99ba8c2e4a5d45da3e4cc0d60664351144ad47
      kind: acceptance
      text: prep activity add accepts session events with started or ended and an optional previous session, and rejects malformed ones, covered by tests
    - id: 39592ba5496443b311aadd646464d1e2ad0a5503a154f13d32df2f718bac7e97
      kind: acceptance
      text: Agents leaves out ended sessions and sessions quiet for 30 minutes, and counts a chain of linked session IDs as one agent, covered by tests
    - id: e4285280c24af4222124888a0a87a27128ace5cff2cecce66480ca4f4e97bcaa
      kind: acceptance
      text: Recently worked on follows linked session IDs, so a resumed session sees its predecessor's issue, covered by a test
    - id: b59e597909546296fed5615c51ec7209578e48af809e7ce4ddf7ccfd212fc954
      kind: acceptance
      text: The Claude Code bridge sends started, ended, and ended plus started-with-previous on an ID change at turn.start
    - id: 014d7dd1cc1ce9c37b260505f0f13508099edf904743beb758bb9136a05b257b
      kind: acceptance
      text: The Agent view's count and tab cycling show only live agents, knowledge describes the lifecycle events, prep check reports no errors, and GitHub CI on main is green
by: claude-code
at: 2026-10-17T09:46:55Z
baseline: 20261017-094421
---
