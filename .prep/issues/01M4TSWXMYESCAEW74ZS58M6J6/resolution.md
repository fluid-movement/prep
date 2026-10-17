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
verification:
  - id: 878b381c19577bf83ef58daf359b424adb8fbfab0a62902a6f83045c38401441
    previous: 34d36477bb238dbeb32e87e86678a74eb054edf93af293409468d59d722bbb34
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:39.781489Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: d0e1336fa37c2adacedd5620f5284819e2690f0fec89d1737ffee9eef94da618
    previous: 878b381c19577bf83ef58daf359b424adb8fbfab0a62902a6f83045c38401441
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Source/knowledge reviewed through prep with peer review; updated Inbox/navigation/editing/mouse/CLI/domain/resolution-review guidance, added bounded Inbox entry and confirmed architecture/design/detail/filter still match. Knowledge-lifecycle remains accurate; it has no scope for confirmation metadata. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:40.783785Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: a424f3f8358b7926062173ea559104002812836f2ddca18fb94aee946bb03308
    previous: d0e1336fa37c2adacedd5620f5284819e2690f0fec89d1737ffee9eef94da618
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:41.772768Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: b47168da2764906073f61721c8b5881616aea5679495747c73e769ec80d31bfd
    previous: a424f3f8358b7926062173ea559104002812836f2ddca18fb94aee946bb03308
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:42.773889Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: c587be31db52bb8efba780991a643127f5ccbc63816fd56d3d4341f8f82f7352
    previous: b47168da2764906073f61721c8b5881616aea5679495747c73e769ec80d31bfd
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:43.782816Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: 3e7ab8ec9891c33457f5ad33a45a854d470456a75558b3f838113efe97f391db
    previous: c587be31db52bb8efba780991a643127f5ccbc63816fd56d3d4341f8f82f7352
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Source/knowledge reviewed through prep with peer review; updated Inbox/navigation/editing/mouse/CLI/domain/resolution-review guidance, added bounded Inbox entry and confirmed architecture/design/detail/filter still match. Knowledge-lifecycle remains accurate; it has no scope for confirmation metadata. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:44.810778Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: f8f18756ea37aceb3f47f67203c171fbc4164018baea25a742b69327db9789d2
    previous: 3e7ab8ec9891c33457f5ad33a45a854d470456a75558b3f838113efe97f391db
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: ff9c78966daa1251fd47c6b419970d745c82ef9a54467218dcc74ba119fd2de2
    result: pass
    method: automated
    command: just ci
    observation: Domain Inbox tests cover all six attention categories, ordering and omission of backlog/dropped work. Inbox key/mouse and 80x24/110x28 goldens pass. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:45.839735Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: f6ab4cc4ff4db1a1ce4cc4c93027f7794116f4c66e23c54b5ea47e62d313f53f
    previous: f8f18756ea37aceb3f47f67203c171fbc4164018baea25a742b69327db9789d2
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 44e647cd474d4aa41f2800f59903f6266d199bd35748e19c00bd23803ba87311
    result: pass
    method: automated
    command: just ci
    observation: TestInboxNotificationsAreNewOnlyAndDoNotStealFocus verifies no initial historical notice, new resolution notice, unchanged selection and no duplicate load notice. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:46.867443Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: ac41f11fe04c2f221922e00d4ddb39d1f0e21a1deeb628188ae107d01177f8f8
    previous: f6ab4cc4ff4db1a1ce4cc4c93027f7794116f4c66e23c54b5ea47e62d313f53f
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 9b8d3a176f0cd5407a2e05e57de11438e8d97170f862b755443ad652915fc05b
    result: pass
    method: automated
    command: just ci
    observation: Review document tests inspect ready requirement baseline, separate current checklist, contract/DoD, reported stale/missing verification and knowledge assessments, exceptions and captured scopes/fingerprint. Adapter Git display tests pass; real terminal bundle rendering/scrolling checked. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:47.904477Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: 0d1b70a3371af078a29f7bad6283d8003abf7a3bbdffd9dbdde0ccce752d00c9
    previous: ac41f11fe04c2f221922e00d4ddb39d1f0e21a1deeb628188ae107d01177f8f8
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: a157cfdb1227b2362e93182f78a759c444785483fc1b3d4460e30d1350407809
    result: pass
    method: automated
    command: just ci
    observation: UI tests retain captured approval through reload, prohibit implicit approval, require legacy scope refresh and handle duplicate/late submissions. Shared domain adapter tests refuse changed code/config scopes, resolution, requirement, policy/contract, knowledge applicability or another review without appending approval. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:48.947657Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: 5efbab532c76f0508179912336c6f36f549b9ee9681dfb8ab09b34f2b752d74b
    previous: 0d1b70a3371af078a29f7bad6283d8003abf7a3bbdffd9dbdde0ccce752d00c9
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 8307e0462ad8bfa7cc7150c1c02f96b95d3bcee0a92613530a310c99e987f188
    result: pass
    method: automated
    command: just ci
    observation: Inbox copyTarget returns selected issue ID for OSC52; key/mouse tests preserve pending review on copy/open/leave. Unknown dependency reasons name unavailable external references; review rendering explicitly labels stale/missing reported artifacts. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:49.998129Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
  - id: 9dabff400d84aa4b434c11238d2b458eb35fa1722551cffbe2fd70c5b5adb70f
    previous: 5efbab532c76f0508179912336c6f36f549b9ee9681dfb8ab09b34f2b752d74b
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 1b0d211bf6fe2b1c01405c1f7411201c4b36875a72ffc408742fd117209aa9d7
    result: pass
    method: automated
    command: just ci
    observation: 'Focused Inbox/adapter tests and final just ci pass: 80x24/110x28/single-pane bounds, four list/bundle goldens, notifications, keyboard/mouse/back/wheel, legacy scopes and stale approval refusal. Eight entries updated and five related entries reviewed unchanged. Rechecked after final attention-caption/golden fix; final just ci and staged diff whitespace checks passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:51.050133Z
    scopes:
      - internal/domain
      - internal/tui
      - internal/cli
      - README.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 04f598c509b58f9c4872eb5f6d03c6f0314cb2959dde9deb9aa598c21e7e6614
knowledge_reviews:
  - id: 5ad75873d9542fb47330c0f42aa9c7c4719a9f41140d9008a4e49e407898a1c4
    previous: 522304ec28f5ca07c16b80bcf47effcf5c80687f5d6a91b53cc19f000e749379
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/tui-inbox.md
    entry_revision: 64ca54bdc9e6af0675f14985c2e0551ff6d664d43b7f1e0c50824cb530e311ee
    observation: Documented and compared classifier precedence, review bundle, notification focus, captured adapter checks and focused UI/adapter tests with source. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:34Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: d8a3b6e25e74a2ef9b7c4ca985041fe8cf31c50075e5e5bdf8a447cdc0d79351
    previous: 5ad75873d9542fb47330c0f42aa9c7c4719a9f41140d9008a4e49e407898a1c4
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/tui.md
    entry_revision: 2fffab59789fdf86f29184a981f2153c85f3a5d1363a91e5e3a6d1b120b4b119
    observation: Added Inbox navigation and linked its bounded component entry; updated scope callbacks, layout, keys and test descriptions. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:38Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: c2782cfcd3babe98db64e9f7e0e10385cc7eefe19cd8a4b15c23d268891241a3
    previous: d8a3b6e25e74a2ef9b7c4ca985041fe8cf31c50075e5e5bdf8a447cdc0d79351
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/tui-navigation.md
    entry_revision: 0c13d5e2ae090ecf9958010e5035982ecf0a5c0cd53d36ece8c731720f455735
    observation: Compared four-screen bar, u key, Inbox list/detail escape and back-stack focus with nav.go and Inbox mouse/key tests. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:42Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: ac001ece8e42652340630181a3eccfa9eef35becf795745540256ce91bfba6d6
    previous: c2782cfcd3babe98db64e9f7e0e10385cc7eefe19cd8a4b15c23d268891241a3
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/tui-editing.md
    entry_revision: d7aca523e99db99eea506086cb2ca441405794d8ca9c0128cd6420ff9f8329c5
    observation: Documented specialized review callback, explicit approval/scope dialogs, stale rejection, duplicate submission and late-response handling; ordinary edits retain shared pipeline. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:46Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: e65a20c63e71c7deabe186958a8f9e9341d37507b74149768ca560cfde02d00b
    previous: ac001ece8e42652340630181a3eccfa9eef35becf795745540256ce91bfba6d6
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/tui-mouse.md
    entry_revision: 217c7a73cd19d7be3622a6f4abd93fa2e79114fc8a7f1360a6b63d6e201f4359
    observation: Documented Inbox pane/row marks and wheel behavior; checked focused mouse row/detail/issue/back tests. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:51Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: eca0da956b8822e08cada714f6dd06571970fca7ce0fde7e8ad261d2cb3c5ef1
    previous: e65a20c63e71c7deabe186958a8f9e9341d37507b74149768ca560cfde02d00b
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/cli.md
    entry_revision: 8f0ea333221662bf55c59a3e8e3d6d2a073d565191800f40b6acbd2ec5a6aaa1
    observation: Linked new adapter behavior and confirmed CLI remains owner of fresh stores and serialized Plan/CheckWrite/Apply with staging, no commits. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:54Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: ef9c5da1f68348eeb74b9134c194b5645330190dd75b9cd89294abd5e43834b7
    previous: eca0da956b8822e08cada714f6dd06571970fca7ce0fde7e8ad261d2cb3c5ef1
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /components/domain.md
    entry_revision: 1225a093187a6a18c28018885441afcd6e5c5c38d8c818ab2c7faa24bdd9985a
    observation: Added Tree.Inbox classifier pointer; read-only categorization and ordering reuse shared domain APIs without changing dependency or completion gates. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:31:58Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
  - id: 78aec4191a7aad53f93391b0c7b94beea287d16ef100e5417f342c3996cd0726
    previous: ef9c5da1f68348eeb74b9134c194b5645330190dd75b9cd89294abd5e43834b7
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    kind: updated
    entry: /features/resolution-review.md
    entry_revision: b7f50cd7652537dbc94ced999babda75c396d35c22f02d46f9097eefd49e345a
    observation: Documented displayed snapshot checks including current contract, requirement/checklist/baseline, status, knowledge applicability and scoped fingerprints; legacy extra scope refresh and pending leaving semantics. Final Inbox attention-count caption and regenerated goldens rechecked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:32:02Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/tui
    snapshot: bc03deb02b82a8624c7a6ff6c5243c446aab71dc1df5930d7462636b1498cef6
outcome: done
by: codex
at: 2026-10-14T17:33:15Z
documentation:
  entries:
    - /components/tui-inbox.md
    - /components/tui.md
    - /components/tui-navigation.md
    - /components/tui-editing.md
    - /components/tui-mouse.md
    - /components/cli.md
    - /components/domain.md
    - /features/resolution-review.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---

Implemented top-level Inbox and bundled completion review with keyboard/mouse navigation, non-blocking new-completion notices, explicit legacy scope refresh and captured-snapshot approval through shared domain actions. Peer review, focused UI/adapter/domain tests, real-terminal navigation/scrolling, final just ci, zero-error/zero-warning prep check, fmt check and staged whitespace check passed. Human review remains pending.
