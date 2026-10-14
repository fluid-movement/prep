---
evidence:
  - id: f3f89a7da3914e83ce1b1f2bca15fb8be41d2f2ed2aecd300edf627cb3d82894
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:44.475596Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: ae2d8ef8ec570a3fc92c2989e33d4a1545dce183552eecbd8b56fa2c900e9229
    previous: f3f89a7da3914e83ce1b1f2bca15fb8be41d2f2ed2aecd300edf627cb3d82894
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Source/knowledge reviewed through prep with peer review; updated Inbox/navigation/editing/mouse/CLI/domain/resolution-review guidance, added bounded Inbox entry and confirmed architecture/design/detail/filter still match. Knowledge-lifecycle remains accurate; it has no scope for confirmation metadata.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:45.334374Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: ff0defe0417415d92877c48a4a6498530e45474670204c3b25ff13b0ff241e18
    previous: ae2d8ef8ec570a3fc92c2989e33d4a1545dce183552eecbd8b56fa2c900e9229
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:46.200965Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: a1e8e13a390aca8df5077eb1c494c176b3f2643aa4bb6b79b1d5ad27cf7c3c64
    previous: ff0defe0417415d92877c48a4a6498530e45474670204c3b25ff13b0ff241e18
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:47.085728Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: 1019feb5672d758c4be8e4a74f289b01f4557721ccffa0fc3415cd57945fc969
    previous: a1e8e13a390aca8df5077eb1c494c176b3f2643aa4bb6b79b1d5ad27cf7c3c64
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: just ci
    observation: 'Final just ci passed: gofmt, go vet ./..., go test ./..., current-source prep check with 0 errors/0 warnings, and prep fmt --check. Focused Inbox/adapter tests and real-terminal navigation/scrolling also passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:47.977118Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: b1242f37c24b245d5b009074213b0a0fdf97556c715f772286a832c3078ed62d
    previous: 1019feb5672d758c4be8e4a74f289b01f4557721ccffa0fc3415cd57945fc969
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Source/knowledge reviewed through prep with peer review; updated Inbox/navigation/editing/mouse/CLI/domain/resolution-review guidance, added bounded Inbox entry and confirmed architecture/design/detail/filter still match. Knowledge-lifecycle remains accurate; it has no scope for confirmation metadata.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:48.867856Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: efaa441fde56725d06bb2725fe707c291a709d6f1091f2a40528d7c897cd17bf
    previous: b1242f37c24b245d5b009074213b0a0fdf97556c715f772286a832c3078ed62d
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: ff9c78966daa1251fd47c6b419970d745c82ef9a54467218dcc74ba119fd2de2
    result: pass
    method: automated
    command: just ci
    observation: Domain Inbox tests cover all six attention categories, ordering and omission of backlog/dropped work. Inbox key/mouse and 80x24/110x28 goldens pass.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:49.77006Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: a46102c2877066599535d6e7f46b48553e31ac2d06b03231a242918a33f330d1
    previous: efaa441fde56725d06bb2725fe707c291a709d6f1091f2a40528d7c897cd17bf
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 44e647cd474d4aa41f2800f59903f6266d199bd35748e19c00bd23803ba87311
    result: pass
    method: automated
    command: just ci
    observation: TestInboxNotificationsAreNewOnlyAndDoNotStealFocus verifies no initial historical notice, new resolution notice, unchanged selection and no duplicate load notice.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:50.675805Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: c75c7239952cbd41e84304542f8676f03b9a06bd0119cb194c029a2d26e37cf2
    previous: a46102c2877066599535d6e7f46b48553e31ac2d06b03231a242918a33f330d1
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 9b8d3a176f0cd5407a2e05e57de11438e8d97170f862b755443ad652915fc05b
    result: pass
    method: automated
    command: just ci
    observation: Review document tests inspect ready requirement baseline, separate current checklist, contract/DoD, reported stale/missing verification and knowledge assessments, exceptions and captured scopes/fingerprint. Adapter Git display tests pass; real terminal bundle rendering/scrolling checked.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:51.590296Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: 5a94d2dd511b03c1b04faafc640c2d45f9d71be1b254cf92ae38427f9deb73ee
    previous: c75c7239952cbd41e84304542f8676f03b9a06bd0119cb194c029a2d26e37cf2
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: a157cfdb1227b2362e93182f78a759c444785483fc1b3d4460e30d1350407809
    result: pass
    method: automated
    command: just ci
    observation: UI tests retain captured approval through reload, prohibit implicit approval, require legacy scope refresh and handle duplicate/late submissions. Shared domain adapter tests refuse changed code/config scopes, resolution, requirement, policy/contract, knowledge applicability or another review without appending approval.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:52.507498Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: 73f86553f5feda30cba8f9c7acf29c5b452e64656887a8f745ecca5e0b84c40e
    previous: 5a94d2dd511b03c1b04faafc640c2d45f9d71be1b254cf92ae38427f9deb73ee
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 8307e0462ad8bfa7cc7150c1c02f96b95d3bcee0a92613530a310c99e987f188
    result: pass
    method: automated
    command: just ci
    observation: Inbox copyTarget returns selected issue ID for OSC52; key/mouse tests preserve pending review on copy/open/leave. Unknown dependency reasons name unavailable external references; review rendering explicitly labels stale/missing reported artifacts.
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:53.431081Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
  - id: 34d36477bb238dbeb32e87e86678a74eb054edf93af293409468d59d722bbb34
    previous: 73f86553f5feda30cba8f9c7acf29c5b452e64656887a8f745ecca5e0b84c40e
    contract: 7283eca8aaaf44202f783e501b12ab92abbce71b42daf2fd8b7bd6d7197d1736
    requirement: 1b0d211bf6fe2b1c01405c1f7411201c4b36875a72ffc408742fd117209aa9d7
    result: pass
    method: automated
    command: just ci
    observation: 'Focused Inbox/adapter tests and final just ci pass: 80x24/110x28/single-pane bounds, four list/bundle goldens, notifications, keyboard/mouse/back/wheel, legacy scopes and stale approval refusal. Eight entries updated and five related entries reviewed unchanged.'
    provenance: reported
    by: codex
    at: 2026-10-14T17:29:54.358656Z
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
    snapshot: f636f12d0d0d800cace30c6343af3be808549be8f6850f2dfd2826255caffd82
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
---
