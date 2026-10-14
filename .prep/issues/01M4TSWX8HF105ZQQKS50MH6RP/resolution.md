---
contract:
  id: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
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
    - id: 8d7187b9b5902802fbe7bce9176977f7534c98ced4021ad7f48349ad20a1d922
      kind: acceptance
      text: Projects can explicitly record verification and documentation policies via CLI; readiness requires a configured policy without blocking capture or definition.
    - id: 8c2ac8d8778027784c23f466465e00276254de6446c97f0e93db2fe71a545b44
      kind: acceptance
      text: Effective completion requirements are bound at readiness and changes require explicit reconciliation.
    - id: 6dd544ee75ae15fcae11c7d5414021eed43f029f3daa40cb3bcfb14073d13b0e
      kind: acceptance
      text: Completion validates applicable evidence per requirement and human-approved exceptions without executing checks or requiring a commit.
    - id: 0d1f06e18ce75d2f1b185a1684087d450655c1a1ea3f1c61128e4d65cb3ac549
      kind: acceptance
      text: Legacy records remain readable without fabricated evidence; domain, storage, CLI and project checks pass.
verification:
  - id: ef51c8e43c4653fbc9617d4f5f6f72d22b1d8a8bd761e29ffe11ab7dac039c45
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: 'just ci passed: gofmt check, go vet ./..., go test ./... including CLI/domain/mdstore approval and legacy compatibility tests, source prep check and fmt check. Later knowledge/record edits rechecked with source CLI: 0 errors, fmt passes.'
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:48.87641Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: 73da068b4b1e13cc514c0a13dbd3c8b6393380a9e1fbb97afcfbe445baab40da
    previous: ef51c8e43c4653fbc9617d4f5f6f72d22b1d8a8bd761e29ffe11ab7dac039c45
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Reviewed architecture, domain, CLI, Markdown store, storage format, lifecycle, completion policy and new completion approvals against current code; updated behavior and documented revision/CAS/manual-edit limits. Reviewed unchanged agent-context, Claude integration and tracking; confirmed them through prep. No claim of authenticated identity or semantic automation.
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:49.175062Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: 0df714c54304dfe4d1e69964f3c65f14ab35fb983edcd60ef5bcb64cd0b467f5
    previous: 73da068b4b1e13cc514c0a13dbd3c8b6393380a9e1fbb97afcfbe445baab40da
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: 'just ci passed: gofmt check, go vet ./..., go test ./... including CLI/domain/mdstore approval and legacy compatibility tests, source prep check and fmt check. Later knowledge/record edits rechecked with source CLI: 0 errors, fmt passes.'
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:49.486026Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: 6b3e2aa70634722072b9987018a61c47fe16646cc88ead9458b82547ce8bb4b0
    previous: 0df714c54304dfe4d1e69964f3c65f14ab35fb983edcd60ef5bcb64cd0b467f5
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: /private/tmp/prep-lifecycle check
    observation: 'Source-built prep check passed on current records: 0 errors and 2 known freshness warnings (TUI and cross-project).'
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:49.80617Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: 7b9d930206ddcfbdd99d50fccf05af9fa10c221660d80637aff6f2eb9877953d
    previous: 6b3e2aa70634722072b9987018a61c47fe16646cc88ead9458b82547ce8bb4b0
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: /private/tmp/prep-lifecycle fmt --check
    observation: Source-built prep fmt --check passed after all issue and knowledge edits; git diff --check passed.
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:50.135105Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: a04adc95171bfbad40cd1aa0cefc49100f436051c41933a63dc64abf2d34fa66
    previous: 7b9d930206ddcfbdd99d50fccf05af9fa10c221660d80637aff6f2eb9877953d
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Reviewed architecture, domain, CLI, Markdown store, storage format, lifecycle, completion policy and new completion approvals against current code; updated behavior and documented revision/CAS/manual-edit limits. Reviewed unchanged agent-context, Claude integration and tracking; confirmed them through prep. No claim of authenticated identity or semantic automation.
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:50.473124Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: 95c1ae33956f3c08c1c8456eab3206477e02cdbf031ebefe52baf50bf4228520
    previous: a04adc95171bfbad40cd1aa0cefc49100f436051c41933a63dc64abf2d34fa66
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: 8d7187b9b5902802fbe7bce9176977f7534c98ced4021ad7f48349ad20a1d922
    result: pass
    method: automated
    command: just ci
    observation: Full just ci passed including TestProjectPolicyCommand and shared readiness policy gates; capture/definition remain available without a policy.
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:50.854688Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: eb77a3335a9a5ed84afa94ad578d5fb3db1df063fbec117cd87558e54108a628
    previous: 95c1ae33956f3c08c1c8456eab3206477e02cdbf031ebefe52baf50bf4228520
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: 8c2ac8d8778027784c23f466465e00276254de6446c97f0e93db2fe71a545b44
    result: pass
    method: automated
    command: just ci
    observation: 'Full just ci passed: policy retention/adoption and previous-contract round trips, acceptance/ancestry restrictions and retained inherited-DoD snapshot behavior verified.'
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:51.21327Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: a2eaf65c059c087697dd478ee944956474b9c78154814a86a971862c215762ec
    previous: eb77a3335a9a5ed84afa94ad578d5fb3db1df063fbec117cd87558e54108a628
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: 6dd544ee75ae15fcae11c7d5414021eed43f029f3daa40cb3bcfb14073d13b0e
    result: pass
    method: automated
    command: just ci
    observation: Full just ci passed, including scoped missing/failed/changed-evidence gates, human/reason requirements, work/config/report invalidation, retention/opt-out change-away-and-back, waived-but-unchecked criteria, resolution approval snapshots and no-commit completion tests.
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:51.584159Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
  - id: ad00930684b65e54c339054a62795801f188af40e8b00090658ecaf6641eab7a
    previous: a2eaf65c059c087697dd478ee944956474b9c78154814a86a971862c215762ec
    contract: 1da0f73317ca04967621e54f26e8069d616a6692645f0c47586a9a411d61f413
    requirement: 0d1f06e18ce75d2f1b185a1684087d450655c1a1ea3f1c61128e4d65cb3ac549
    result: pass
    method: automated
    command: just ci
    observation: 'just ci passed: gofmt check, go vet ./..., go test ./... including CLI/domain/mdstore approval and legacy compatibility tests, source prep check and fmt check. Later knowledge/record edits rechecked with source CLI: 0 errors, fmt passes.'
    provenance: reported
    by: codex
    at: 2026-10-14T12:12:51.961788Z
    scopes:
      - internal
      - cmd
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - justfile
      - .github/workflows
    snapshot: d1fe998846ed5c9894eaa8e98767ea1e0f98a2a6bcebee325f2f80cec9d40543
outcome: done
by: codex
at: 2026-10-14T12:13:23Z
documentation:
  entries:
    - /features/completion-policy.md
    - /features/completion-approvals.md
    - /features/lifecycle.md
    - /components/architecture.md
    - /components/domain.md
    - /components/cli.md
    - /components/markdown-store.md
    - /conventions/storage-format.md
    - /features/agent-context.md
    - /components/claude-code.md
    - /decisions/tracking.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---
