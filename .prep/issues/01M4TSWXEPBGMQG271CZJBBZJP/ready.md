---
contract:
  id: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
  verification: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
  documentation: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
  ancestors:
    - 01M4TST3C4ZSZDKA5WKBA3ND4H
    - 01M498VPZAAQJKHXGZZ3PWJH2F
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
    - id: 0d7319b793c1cbfa3a7a266fbaa1da264d6c89cec8609de464659e5489225b6d
      kind: acceptance
      text: Implementation completion and pending/approved human review are distinct and visible; pending review does not block unrelated ready issues.
    - id: 0e69efe0a01e1b6603d42f95697cd0865b1a1d7bee1d6acf05fbeb79d392a060
      kind: acceptance
      text: A human-attributed approval identifies a specific resolution and code/evidence snapshot without requiring a commit; agent-attributed approval is rejected.
    - id: 711887acd68359627a941cbd1565066b46f4a0c1d867fa998bb50971fc10328a
      kind: acceptance
      text: Changing the reviewed snapshot or completing a reopened issue cannot reuse a prior approval, including change-away-and-back cases.
    - id: 80dae0d9c850987cb846c45758047f8c11d5c612df0d6e2182d6da98a6a6e8d0
      kind: acceptance
      text: Explicit reopening preserves prior resolutions, verification and review history and resumes the same issue through shared domain/CLI transitions; feedback alone does not reopen.
    - id: adcf746472c99a531190411bbdbee32b7dd16462a5798f4991d226e73731aba9
      kind: acceptance
      text: Dependency satisfaction defaults to implementation completion and supports an explicit human-approval prerequisite with consistent domain/CLI behavior.
    - id: 1e2f698fdb9f760e228e015d7f3ac2f316a6c5b4906680382f39a0557e89e8a8
      kind: acceptance
      text: Versioned records round-trip and validate without fabricating approvals for legacy issues; regression coverage exercises approval, invalidation, reopen history and dependency gates.
by: codex
at: 2026-10-14T12:18:18Z
baseline: 20261014-115939
---
