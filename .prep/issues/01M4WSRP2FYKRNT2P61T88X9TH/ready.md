---
contract:
  id: 3f23421701707c46ad81dce75fd0dc3e0f499f2dc8d11dffff6d6a8f23b13a45
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
    - id: c9ed79ef5af5e11a02bae38eb71b33a4c76c70b4c10f935ed45baf2acd160582
      kind: acceptance
      text: Enrichment and implementation guidance offers bounded, relevant knowledge candidates with targeted prep knowledge find/show commands; candidates do not imply semantic correctness.
    - id: 769d646bdde3aee5a53545cefceec5d3902f86eee7c67804a551ca22acebf7b2
      kind: acceptance
      text: Completion requires per-work evidence of affected knowledge updates or explicit unchanged reviews, or a justified no-impact assessment; merely listing --docs entries cannot satisfy the gate.
    - id: 4d5ed4708acfc3596a2215e8d189932f9187890ae2288ee55a8700724795e505
      kind: acceptance
      text: Knowledge review records preserve actor, review/update kind and the issue/work and entry revision they apply to; relevant later changes invalidate applicability without fabricating reviews.
    - id: d747f61b030670a8713d74e041d0318a58611858b0430661f9342fa79909c953
      kind: acceptance
      text: CLI and TUI use the same domain completion gate, and legacy records remain readable without invented knowledge evidence.
    - id: 40c8e10010cb77820f90466844df3973acb379455fc69175c36067a90a659b61
      kind: acceptance
      text: Regression coverage includes missing review, stale review, updated entries, unchanged confirmation, justified no-impact and bounded candidate output; affected knowledge is maintained.
by: codex
at: 2026-10-14T12:16:55Z
baseline: 20261014-115939
---
