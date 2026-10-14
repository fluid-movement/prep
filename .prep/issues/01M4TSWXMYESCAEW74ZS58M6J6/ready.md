---
contract:
  id: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
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
    - id: ff9c78966daa1251fd47c6b419970d745c82ef9a54467218dcc74ba119fd2de2
      kind: acceptance
      text: Inbox is a top-level screen beside Issues, Agent and Knowledge, prioritizing review/sign-off/questions/staleness/dependency uncertainty while ordinary backlog stays in Issues.
    - id: 44e647cd474d4aa41f2800f59903f6266d199bd35748e19c00bd23803ba87311
      kind: acceptance
      text: A completed issue becomes available for review in Inbox and emits a non-blocking notification without changing the user's focus.
    - id: 9b8d3a176f0cd5407a2e05e57de11438e8d97170f862b755443ad652915fc05b
      kind: acceptance
      text: A bundled review displays requirements, identifiable code changes/snapshot, acceptance/DoD results, knowledge evidence and approved exceptions from the versioned contract.
    - id: a157cfdb1227b2362e93182f78a759c444785483fc1b3d4460e30d1350407809
      kind: acceptance
      text: Human approval uses the shared domain review action for the displayed snapshot; leaving review keeps it pending, and there is no reject/request-changes or implicit reopen action.
    - id: 8307e0462ad8bfa7cc7150c1c02f96b95d3bcee0a92613530a310c99e987f188
      kind: acceptance
      text: Issue IDs can be copied for feedback; stale/missing artifacts and unresolved dependency information are visible rather than presented as successful verification.
    - id: 1b0d211bf6fe2b1c01405c1f7411201c4b36875a72ffc408742fd117209aa9d7
      kind: acceptance
      text: Navigation, rendering, notification and stale-approval behavior have focused TUI tests; legacy issues render safely and affected knowledge is maintained.
by: codex
at: 2026-10-14T16:58:35Z
baseline: 20261014-120009
---
