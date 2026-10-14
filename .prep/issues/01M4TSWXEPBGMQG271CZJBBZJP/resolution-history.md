---
resolutions:
  - record:
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
      verification:
        - id: 71eac66bc2ff9b01a5572c452f0b1ee87fa8b2888cb38b48156b9eab557fb5c5
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:38.70191Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 2c24fdebbdaaff668e04be389abf32532adb7f29b076d28c0939a249a1d9d751
          previous: 71eac66bc2ff9b01a5572c452f0b1ee87fa8b2888cb38b48156b9eab557fb5c5
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
          result: pass
          method: manual
          observation: Reviewed and maintained resolution-review, knowledge-lifecycle, lifecycle, agent-context, architecture, domain, CLI, Markdown store, TUI editing and storage-format entries. Explicit current knowledge reports recorded; reviewed remaining affected integration/tracking entries as unchanged. Reports are attributed assessments, not authenticated approval.
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:39.251428Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 649970a60856317701d4c046b692ad704f8788079a9b59b650d4d4b7fc04fe66
          previous: 2c24fdebbdaaff668e04be389abf32532adb7f29b076d28c0939a249a1d9d751
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:39.814874Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 9e1f3bced30238e03c5d4df837603bd46fdf0492c507a5cffd09670c624d120a
          previous: 649970a60856317701d4c046b692ad704f8788079a9b59b650d4d4b7fc04fe66
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:40.383193Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 29334c2a126e71a1214e6909bc1a928504235f82850abd1da8aea2e64e5d0004
          previous: 9e1f3bced30238e03c5d4df837603bd46fdf0492c507a5cffd09670c624d120a
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:40.963981Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: b181739780ceff20693fae57d87f5db8b866e78bbfb28bf865f428ca668ac24b
          previous: 29334c2a126e71a1214e6909bc1a928504235f82850abd1da8aea2e64e5d0004
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
          result: pass
          method: manual
          observation: Reviewed and maintained resolution-review, knowledge-lifecycle, lifecycle, agent-context, architecture, domain, CLI, Markdown store, TUI editing and storage-format entries. Explicit current knowledge reports recorded; reviewed remaining affected integration/tracking entries as unchanged. Reports are attributed assessments, not authenticated approval.
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:41.553337Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: f1799af8c7da3deb8777270594b726f47dd103d197f6a7ea83aa5cf7d5906d30
          previous: b181739780ceff20693fae57d87f5db8b866e78bbfb28bf865f428ca668ac24b
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 0d7319b793c1cbfa3a7a266fbaa1da264d6c89cec8609de464659e5489225b6d
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:42.148858Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 3a53b174929d423951991c302cec60b6be80882b4243b2942a5b071dbd129349
          previous: f1799af8c7da3deb8777270594b726f47dd103d197f6a7ea83aa5cf7d5906d30
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 0e69efe0a01e1b6603d42f95697cd0865b1a1d7bee1d6acf05fbeb79d392a060
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:42.753811Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 6f51fb4de6e7ec572fc5a82d00f098cc67d8020565a791bd046604e9756b179d
          previous: 3a53b174929d423951991c302cec60b6be80882b4243b2942a5b071dbd129349
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 711887acd68359627a941cbd1565066b46f4a0c1d867fa998bb50971fc10328a
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:43.371266Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: df8292f37adcc98d6289c756d752a288d38b28074065e1f12c80459733f5d38e
          previous: 6f51fb4de6e7ec572fc5a82d00f098cc67d8020565a791bd046604e9756b179d
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 80dae0d9c850987cb846c45758047f8c11d5c612df0d6e2182d6da98a6a6e8d0
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:43.997953Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 82f6bb1d72f2cab722fa873d830738865c87997dac6e495a374b9504fc64d3de
          previous: df8292f37adcc98d6289c756d752a288d38b28074065e1f12c80459733f5d38e
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: adcf746472c99a531190411bbdbee32b7dd16462a5798f4991d226e73731aba9
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:44.630975Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: f809078f71a35f79eb1785ca45bc43bd8d50f998bf3ad5059ed14962bb864bf2
          previous: 82f6bb1d72f2cab722fa873d830738865c87997dac6e495a374b9504fc64d3de
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          requirement: 1e2f698fdb9f760e228e015d7f3ac2f316a6c5b4906680382f39a0557e89e8a8
          result: pass
          method: automated
          command: just ci
          observation: 'Integrated just ci passed: gofmt, go vet ./..., all package tests, prep check 0 errors/0 warnings and fmt check. Focused combined regression proves knowledge work/config/entry changes invalidate review and reopen preserves selected reports; review tests cover human/reason/scope gates, dependency approval, legacy history and approval before commit remaining approved through show/guide/review after unchanged commit. No actual human approval created for this project.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:45.272883Z
          scopes:
            - internal
            - cmd
            - .prep/knowledge
          config_scopes:
            - go.mod
            - go.sum
            - justfile
            - .github/workflows
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
      knowledge_reviews:
        - id: 37d765e4ee9c1209203011eeb88c6666fd89ba2e3b54632bf669250bbee6e8b5
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /features/resolution-review.md
          entry_revision: d2c6b4d77cf2d51d372b6ce8a1f7bb29cb8f6fcabc304c7b1a3bcde054b80387
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:33Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 0b94b8241801af78229e5c5fdcf01284fc19e6ffe1f58e1c7e52c123ca608b5d
          previous: 37d765e4ee9c1209203011eeb88c6666fd89ba2e3b54632bf669250bbee6e8b5
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: unchanged
          entry: /features/knowledge-lifecycle.md
          entry_revision: 4131a4c657c04d5055fdd049eac09c6d81c5bbdedb24f7e512e9d90564326723
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:34Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: d4b27b49e38848a56e53239749191fd0bac51d3a18ee252306d7f3b58f8e0edb
          previous: 0b94b8241801af78229e5c5fdcf01284fc19e6ffe1f58e1c7e52c123ca608b5d
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /features/lifecycle.md
          entry_revision: 7b421850ce5a71d864c12c3fe4a4f239e86532c3ad895015705ec9bdc0b159aa
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:34Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 88b622b7b4c3c5077c72fdd321651789dba601dd1ba05ad93db64ad464358b9b
          previous: d4b27b49e38848a56e53239749191fd0bac51d3a18ee252306d7f3b58f8e0edb
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /features/agent-context.md
          entry_revision: 599db229f76c018b76c8aa2070aa05bd74be1201432c691526592b2a25e1aa0c
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:35Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 83a06bb7ff11b239aa60804b620ffa170b5c41696b974f15094dea8bdbf420f1
          previous: 88b622b7b4c3c5077c72fdd321651789dba601dd1ba05ad93db64ad464358b9b
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /components/architecture.md
          entry_revision: f746af8e3c8a0ea2d70d69417420e310030fa58ac14c442a9192f7376e698bda
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:35Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 4cb8a3510b6348aa5ee15b0c5f176425a06917b5a3b6cad4229682335d50124f
          previous: 83a06bb7ff11b239aa60804b620ffa170b5c41696b974f15094dea8bdbf420f1
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /components/domain.md
          entry_revision: b75892b1382ca16eef74943de9543b6f70888669d0a8494397156ac367364d9f
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:36Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 37be5fdeb11007d2ddc02ea98ca43525fca3fe9aa3c7a57286e80b49f45af763
          previous: 4cb8a3510b6348aa5ee15b0c5f176425a06917b5a3b6cad4229682335d50124f
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /components/cli.md
          entry_revision: a7f97e68ff1e62243fad408f3c7695200ce6f4315cc4e7519d29060e6072c5b1
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:36Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 361d365b52c4a5dde978b5a9c188f32d548c2fc8c2be4dd4ee8e0a5ce10758f2
          previous: 37be5fdeb11007d2ddc02ea98ca43525fca3fe9aa3c7a57286e80b49f45af763
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /components/markdown-store.md
          entry_revision: 5dff84053ea66ae72307ab37f9361da0a6ba58aa2fe719dde46b7eab64acde71
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:37Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: e9a070a989e96fc75a24d8b73fdf57bb2a5111ca41e3669fe24b830af2b1ac21
          previous: 361d365b52c4a5dde978b5a9c188f32d548c2fc8c2be4dd4ee8e0a5ce10758f2
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /components/tui-editing.md
          entry_revision: 57113e56e61925015124b51b400eb8775f74fbfae4c4e361baa92708f11604ad
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:37Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
        - id: 5b9283833b9f7f49b5f274bdcbffd156a607a754d6f6b2604e384598e4277d90
          previous: e9a070a989e96fc75a24d8b73fdf57bb2a5111ca41e3669fe24b830af2b1ac21
          contract: ebdacd74ed4299428487fd2eb9f0bc497d8bd0975e9576ff4ece6b9b49df43a7
          kind: updated
          entry: /conventions/storage-format.md
          entry_revision: 024b06a4c5cf73d7d5db781fdf2a43d0977a181f17e78f6a92b1b1b5eafbc639
          observation: 'Reviewed against the integrated implementation: versioned human review/reopen, full resolution identities, selected knowledge work/config and entry-file coverage, shared gates, legacy behavior and presentation-only Git evidence. Updated this entry where behavior changed; knowledge-lifecycle entry still matches.'
          provenance: reported
          by: codex
          at: 2026-10-14T16:32:38Z
          scopes:
            - .prep/knowledge
            - cmd
            - internal
          config_scopes:
            - .github/workflows
            - go.mod
            - go.sum
            - justfile
          snapshot: b6689251bcfe7fce205a0a059ee1e41512184f09641053fe1dede28e1e2e7381
      generation: 1
      outcome: done
      by: codex
      at: 2026-10-14T16:32:46Z
      documentation:
        entries:
          - /features/resolution-review.md
          - /features/knowledge-lifecycle.md
          - /features/lifecycle.md
          - /features/agent-context.md
          - /components/architecture.md
          - /components/domain.md
          - /components/cli.md
          - /components/markdown-store.md
          - /components/tui-editing.md
          - /conventions/storage-format.md
      dod:
        - go test ./... passes
        - prep check reports no errors
        - prep fmt --check passes
        - knowledge entries describing changed behavior are updated
---
