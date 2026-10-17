---
contract:
  id: 347c9db7c9e740963157a882bf7c7632fc1560a8bb3ca18abb4139752e45336a
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
    - id: 0376af7a0f541250c4071193c2e9cdbd7b6e9709a0d6753330554a9d0cdabfa2
      kind: acceptance
      text: Record the observed binary version, project/worktree, TUI screen and filter, and whether restart corrects the display; capture limits if the original setup is unavailable.
    - id: bf0279228658c9de1ef4626e821c479122bad474a701cd5c3469313f8bdc5ec4
      kind: acceptance
      text: Establish whether external lifecycle changes produce watcher events and whether the TUI reloads and applies the resulting issue state.
    - id: ffd3b0445b12a9fe22826dd84dad09c3b4e5450b7b1231dbc6ce568a6b31e172
      kind: acceptance
      text: Document findings and a focused regression/fix proposal only if a defect is confirmed, including any overlap with TUI Inbox work.
previous_contracts:
  - id: 3d510db68bfc168754cb62b81d20cf4fd85573bc89a5413ba32d5b06bb9473e3
    verification: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
    documentation: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
    ancestors:
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
      - id: 0376af7a0f541250c4071193c2e9cdbd7b6e9709a0d6753330554a9d0cdabfa2
        kind: acceptance
        text: Record the observed binary version, project/worktree, TUI screen and filter, and whether restart corrects the display; capture limits if the original setup is unavailable.
      - id: bf0279228658c9de1ef4626e821c479122bad474a701cd5c3469313f8bdc5ec4
        kind: acceptance
        text: Establish whether external lifecycle changes produce watcher events and whether the TUI reloads and applies the resulting issue state.
      - id: ffd3b0445b12a9fe22826dd84dad09c3b4e5450b7b1231dbc6ce568a6b31e172
        kind: acceptance
        text: Document findings and a focused regression/fix proposal only if a defect is confirmed, including any overlap with TUI Inbox work.
by: codex
at: 2026-10-14T12:10:15Z
baseline: 20261014-120851
---
