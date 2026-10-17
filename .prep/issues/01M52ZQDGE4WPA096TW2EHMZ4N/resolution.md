---
contract:
  id: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
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
    - id: a56be183f6b81168e673ad5847d6d693acc15bbafc03ce744c95194f1cb16dd9
      kind: acceptance
      text: prep help lists no review or dependency override command, and no record in this repository has review, approval-prerequisite or override files or fields
    - id: 43e94dd242ec10d36fe9e06b3d617285b6a0ef743bb7cf95b5fa7d66b163b26d
      kind: acceptance
      text: The TUI's screens are Issues, Agent and Knowledge; goldens are updated and no Inbox code remains
    - id: f2249e81f343568bf3ed463351a3348eb0ea024c3de17854b46fbdbb5cef2a32
      kind: acceptance
      text: Reopening a done issue with a reason appends its resolution to history.md and leaves it ready, covered by a test
    - id: e4aecae75bfa3caa38a4b94e548ed21dcb721127f4ddc1f07b8bfcc9c0f28668
      kind: acceptance
      text: git grep finds no resolution review, generation, approval prerequisite or Inbox types
    - id: 7ef3a5405f485ce14e3ea7cd9cdbefd1421cccca1fb8dd554a55de3800311ccd
      kind: acceptance
      text: Knowledge describing review, the Inbox and reopen is deleted or updated, prep check reports no errors, and GitHub CI on main is green
verification:
  - id: c6ff94535a299a6eb1f2b246fb5c774bcdfd28ea2d1f6d733c1542c6b7c1603b
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: go test ./...; prep check; prep fmt --check
    observation: checker ran go test ./... on main after the merge (all ok); GitHub CI run 39020743145 on 0a84695 succeeded
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:22:36.25643Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 6668af946e1c49fdf184be8c92fe24c52cf546a9a1c52a8026cc4aaf96135aa6
    previous: c6ff94535a299a6eb1f2b246fb5c774bcdfd28ea2d1f6d733c1542c6b7c1603b
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: subagent updated 16 entries; checker removed a stale reference to a deleted test in local-write-safety, deleted the resolution-review and tui-inbox entries through the conversion, and confirmed 10 drifted entries after reading them
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:22:38.578083Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 8663e57b88ed9eaf28d3b9fac2d507b05bf4a65e19528e46a16db5371a9dbbc9
    previous: 6668af946e1c49fdf184be8c92fe24c52cf546a9a1c52a8026cc4aaf96135aa6
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: go test ./...
    observation: all packages ok on main after the merge
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:22:42.995744Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 68f2488016d5b973322a8a54c32ce50c0e45e3eb55d18a04635641aa1744008f
    previous: 8663e57b88ed9eaf28d3b9fac2d507b05bf4a65e19528e46a16db5371a9dbbc9
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: prep check
    observation: 0 errors after converting records with convert-records.py review (including four resolutions written on main during the wave), 1 warning (I029)
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:22:47.633595Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: e236267dd101a127166c2ba9756e104926768ce61ca79f9af8dfee16c7e0efd8
    previous: 68f2488016d5b973322a8a54c32ce50c0e45e3eb55d18a04635641aa1744008f
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: prep fmt --check
    observation: exit 0
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:22:52.32591Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 5e5c5f6154b35c52502c22ed8fa9fbe56db1313be11560841e709315ab772474
    previous: e236267dd101a127166c2ba9756e104926768ce61ca79f9af8dfee16c7e0efd8
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: lifecycle, storage-format, cli, domain, tui-navigation and the other updated entries describe the minimal reopen and no review or Inbox
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:22:57.703934Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 4700f5bfcda97f3a779a65a23dc5b7e0abba6f70186ea4cfa70789dede46811f
    previous: 5e5c5f6154b35c52502c22ed8fa9fbe56db1313be11560841e709315ab772474
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: a56be183f6b81168e673ad5847d6d693acc15bbafc03ce744c95194f1cb16dd9
    result: pass
    method: automated
    command: prep --help; convert-records.py review --dry-run
    observation: no review or dependency override command; the dry run finds no review, history or override record files and no generation or depends_on_approval fields
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:02.194207Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 75fe36de70d062f0771f88adfbd3eb28207475fe3ca809e608358cde7e0d2952
    previous: 4700f5bfcda97f3a779a65a23dc5b7e0abba6f70186ea4cfa70789dede46811f
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: 43e94dd242ec10d36fe9e06b3d617285b6a0ef743bb7cf95b5fa7d66b163b26d
    result: pass
    method: automated
    command: go test ./internal/tui/...
    observation: screens are Issues, Agent and Knowledge; goldens updated (screens bar, hint, keys dialog); no Inbox code
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:06.666422Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 9ad9a7262c12afe5f330da37c186111aa5bf39e04d610a01ee47d4cb826efc1b
    previous: 75fe36de70d062f0771f88adfbd3eb28207475fe3ca809e608358cde7e0d2952
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: f2249e81f343568bf3ed463351a3348eb0ea024c3de17854b46fbdbb5cef2a32
    result: pass
    method: automated
    command: go test ./internal/cli -run TestReopenMovesResolutionToHistoryAndReturnsToReady
    observation: reopen without a reason fails; with a reason the resolution moves into history.md, resolution.md is removed and the issue is ready and claimable
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:11.426423Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 5579715446198b125e9853a27c64624aa12f61569cfc3c0eaa129738ffc774d6
    previous: 9ad9a7262c12afe5f330da37c186111aa5bf39e04d610a01ee47d4cb826efc1b
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: e4aecae75bfa3caa38a4b94e548ed21dcb721127f4ddc1f07b8bfcc9c0f28668
    result: pass
    method: automated
    command: git grep -n -E 'ResolutionReview|DependsOnApproval|Inbox|resolution-history' outside issue records
    observation: no matches
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:15.927855Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: fe8592f1a8def6ca4f4cf2ca0eadf0541394b1e68d2e18f9364a86827ec4e690
    previous: 5579715446198b125e9853a27c64624aa12f61569cfc3c0eaa129738ffc774d6
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    requirement: 7ef3a5405f485ce14e3ea7cd9cdbefd1421cccca1fb8dd554a55de3800311ccd
    result: pass
    method: automated
    command: gh run view 39020743145
    observation: knowledge updated or deleted, prep check 0 errors, GitHub CI on 0a84695 succeeded
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:20.378617Z
    scopes:
      - internal/domain
      - internal/mdstore
      - internal/cli
      - internal/tui
      - internal/projects
      - README.md
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
knowledge_reviews:
  - id: a2e8680ca3892c3c5afdb7faa1d5f0cc9dfefbd4a3a1f79cc2968f67a290fb7b
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /features/lifecycle.md
    entry_revision: 7fc3fcaa29290ab95fa67f404942e6f85c83dd01a40e09394bd896595a54a0a1
    observation: reopen is minimal; no review transition
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:24Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: e0d90be3af06126a02f546b92f8f75488c6a9366ac420d7c27432d280fc8d2de
    previous: a2e8680ca3892c3c5afdb7faa1d5f0cc9dfefbd4a3a1f79cc2968f67a290fb7b
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /conventions/storage-format.md
    entry_revision: 4de9067d4c54123714c297ff6613ace48c402d21e8d6d98b9adaf4ca769d1095
    observation: no review, history or override records or fields
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:29Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 5e3aef05a8e7a4c76e58ca8a65247a771ae53163a078846b0c92154b38edcf8a
    previous: e0d90be3af06126a02f546b92f8f75488c6a9366ac420d7c27432d280fc8d2de
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /components/cli.md
    entry_revision: 1dd9f00fac447171f077f468af1c424e8b6a2c4e2c0e6b65bc8e6da34c3644b1
    observation: no prep review or dependency override
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:34Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: d5a1bc02448f66acc33bae53b4c8f6d2a34e1eea028ebe9315f688e79ea03f42
    previous: 5e3aef05a8e7a4c76e58ca8a65247a771ae53163a078846b0c92154b38edcf8a
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /components/domain.md
    entry_revision: a04ee90e1b96c669326a54c3b4e6752faac70169952a06d15ead40782e0e4363
    observation: no review, generation or approval prerequisites
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:38Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 7ec99d6fafe3cc742e53c74ae5fa7b9354add36370568cab1bdbd259fea25367
    previous: d5a1bc02448f66acc33bae53b4c8f6d2a34e1eea028ebe9315f688e79ea03f42
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /components/tui-navigation.md
    entry_revision: 52c9c440a12035ec74429f5e3c0727742d846beb2ba3d09d4687f96aba056b16
    observation: screens are Issues, Agent and Knowledge
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:43Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: 007d5430b7b0b199fb2bbe407c7af0c51c0d06adee6813e6336a47e6fe33a95d
    previous: 7ec99d6fafe3cc742e53c74ae5fa7b9354add36370568cab1bdbd259fea25367
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /components/markdown-store.md
    entry_revision: ed7e3a91725ce1b0cf90144cc22c507866eb000ea0a51cbf208de3c80d51b547
    observation: reopen appends the resolution to history.md
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:47Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
  - id: bf154fa4506b03f9742c133d49284815e8adf2c5ff746daac7e005a491ca11c5
    previous: 007d5430b7b0b199fb2bbe407c7af0c51c0d06adee6813e6336a47e6fe33a95d
    contract: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
    kind: updated
    entry: /features/local-write-safety.md
    entry_revision: 85a4bbdc017de39a6a916735e788f8ecb9384ccf6bb8622159a15c816dd239dc
    observation: removed reference to a deleted override test
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T09:23:52Z
    scopes:
      - README.md
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/projects
      - internal/tui
    snapshot: 592afe749d9f8909d786e47c72e06b8ddb98ee5dec20d0c28dc24165a013bc3b
generation: 1
outcome: done
by: claude-code/opus-5.5
at: 2026-10-17T09:23:57Z
documentation:
  entries:
    - /features/lifecycle.md
    - /conventions/storage-format.md
    - /components/cli.md
    - /components/domain.md
    - /components/tui-navigation.md
    - /components/markdown-store.md
    - /features/local-write-safety.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---
