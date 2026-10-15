---
contract:
  id: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
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
    - id: 35df46c4420873d5d45667ad68570db8b7d485b9121d0f7013b5b53122915ba8
      kind: acceptance
      text: External dependency status distinguishes satisfied, unsatisfied and unknown and exposes source project/checkout/revision plus uncertainty through CLI and shared domain data.
    - id: f5e7ee770c88f36430932a5e04aec0965efad9e350df7691397def7c3fed7dfb
      kind: acceptance
      text: An unavailable source checkout or unresolved external issue remains unknown and blocks claiming by default; cached or missing data cannot imply satisfaction.
    - id: f3e84d34a412c6e354596dc0c834ee4c8545ecc4e10ce38528161c4e7b17c343
      kind: acceptance
      text: A human-attributed override explicitly identifies the uncertain dependency and records a reason and applicable snapshot; agent-only or missing-reason overrides fail.
    - id: f13d344fe7670f5acb3aeb26bc34dd2a9a9bf026172d9a3241d3b0abe2b3f278
      kind: acceptance
      text: Dependencies default to implementation completion and can explicitly require applicable human approval using the review contract rather than a separate approval definition.
    - id: fa928a80a61bbaa5d93185bd7097f4596a46aeeddd835766f8c700a1824f03bb
      kind: acceptance
      text: Source or dependency changes make stale observations/overrides visible and require reconciliation; local legacy dependencies retain compatible behavior.
    - id: fe63da15ad171c4a81ecbdb7e64b4659fb3688138c1cfb7264a146c1e09739a9
      kind: acceptance
      text: Tests cover available/unavailable projects, unresolved targets, done-but-unapproved prerequisites and scoped overrides; no remote synchronization is required.
verification:
  - id: d5dfc2a73d6d1f0a72621f6d8ca6ebaec6ff8b2b415ddcad75f757ee433ebfdc
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:38.90972Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: 836c3bb54608aaebbf1f2b504abefcaba865949465003454e8002702684f5f43
    previous: d5dfc2a73d6d1f0a72621f6d8ca6ebaec6ff8b2b415ddcad75f757ee433ebfdc
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Reviewed and updated affected knowledge through prep, README and synchronized harness skills. Scoped knowledge assessments record changed behavior and unchanged boundaries.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:40.223563Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 9624364e5a8e88be1a44f8a6c00ff336fa1ee0c3986c947e0e0c1410ce761edb
  - id: 0b2de15eaf60ad6dd445e25eacb4ba57658cb73fa0e163b78d3ae64017036801
    previous: 836c3bb54608aaebbf1f2b504abefcaba865949465003454e8002702684f5f43
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:41.552447Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: efe020b28d68b45738d49c5cb2e5f8682c1a932c1ebdfffe7ac426cd77eb9d73
    previous: 0b2de15eaf60ad6dd445e25eacb4ba57658cb73fa0e163b78d3ae64017036801
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:42.887333Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: 042286cc02ed04a47fc30afb3c7943e1f2725cdcf897f5a7d51b92c2fec5b749
    previous: efe020b28d68b45738d49c5cb2e5f8682c1a932c1ebdfffe7ac426cd77eb9d73
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:44.211943Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: fce73edafe9dca47644bcd7c4e4e8a77cf1942ef15b790db507d69f1b7fca879
    previous: 042286cc02ed04a47fc30afb3c7943e1f2725cdcf897f5a7d51b92c2fec5b749
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Reviewed and updated affected knowledge through prep, README and synchronized harness skills. Scoped knowledge assessments record changed behavior and unchanged boundaries.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:45.549318Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 9624364e5a8e88be1a44f8a6c00ff336fa1ee0c3986c947e0e0c1410ce761edb
  - id: 6a0d8f50f245e24b815b7d8a066e71f8259a5bf335a6c3c59a55b4f634048613
    previous: fce73edafe9dca47644bcd7c4e4e8a77cf1942ef15b790db507d69f1b7fca879
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: 35df46c4420873d5d45667ad68570db8b7d485b9121d0f7013b5b53122915ba8
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:47.928099Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: a50ff1478b2dd472defe7597c49ade13914e5eee46d62986c39849a258f73673
    previous: 6a0d8f50f245e24b815b7d8a066e71f8259a5bf335a6c3c59a55b4f634048613
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: f5e7ee770c88f36430932a5e04aec0965efad9e350df7691397def7c3fed7dfb
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:49.312586Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: acc556ed7008db8f33a8d0a3e49b0fd41559c57a0661bcc663d853a0e741a90d
    previous: a50ff1478b2dd472defe7597c49ade13914e5eee46d62986c39849a258f73673
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: f3e84d34a412c6e354596dc0c834ee4c8545ecc4e10ce38528161c4e7b17c343
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:50.691114Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: 38962407fc6e7f81f24acce9923d02e9a60ede25aff4ba51a2e95296c02ea513
    previous: acc556ed7008db8f33a8d0a3e49b0fd41559c57a0661bcc663d853a0e741a90d
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: f13d344fe7670f5acb3aeb26bc34dd2a9a9bf026172d9a3241d3b0abe2b3f278
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:52.086142Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: c6ee83f34b0c0d88a96e3a482980fc610b3bd02b2c307696cffa11d5ba1e9610
    previous: 38962407fc6e7f81f24acce9923d02e9a60ede25aff4ba51a2e95296c02ea513
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: fa928a80a61bbaa5d93185bd7097f4596a46aeeddd835766f8c700a1824f03bb
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:53.472055Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
  - id: 1d16226cd9b02c830bbd0a5896bb50e5bec0e922ba3c3acce327c71b71613ed5
    previous: c6ee83f34b0c0d88a96e3a482980fc610b3bd02b2c307696cffa11d5ba1e9610
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    requirement: fe63da15ad171c4a81ecbdb7e64b4659fb3688138c1cfb7264a146c1e09739a9
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover fresh unavailable/missing/malformed project observations, required applicable human review, exact human overrides and stale/retired reconciliation; no remote synchronization.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:54.850559Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: c326d775064694d1fd56d7693764176a2809a17df7e1427bed672e589f9296a8
knowledge_reviews:
  - id: 73f588e354fde3f4889e39bf9711a4b07839063a8b64de9272ef53523512a256
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    kind: updated
    entry: /features/cross-project.md
    entry_revision: 771c409691a53e2420fdda94a28f307c6b14c989d4bf1b0a6502e99d057d1cc6
    observation: Fresh tri-state source observations, source provenance, explicit human scoped overrides and stale/retired visibility match domain/projects/CLI/TUI. No cached satisfaction or remote sync.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:14Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: 56f02f1e37ad03e16b6d923e2bb4a1be57048b745dc4c71aafac4c4be4f7fdc5
    previous: 73f588e354fde3f4889e39bf9711a4b07839063a8b64de9272ef53523512a256
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    kind: updated
    entry: /features/lifecycle.md
    entry_revision: 9944a69c9082456e87da8390c28822d13a3c1efdf86b1329c995570ef96feefd
    observation: Unknown external dependencies block claiming by default; current overrides affect only implementation uncertainty. Explicit approval prerequisites share versioned resolution review.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:15Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: d21d9b18e28bbf29dd0ec470d0a4ee7f1fe184fee5241f649e027f99fcc6b100
    previous: 56f02f1e37ad03e16b6d923e2bb4a1be57048b745dc4c71aafac4c4be4f7fdc5
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    kind: updated
    entry: /features/resolution-review.md
    entry_revision: 82eeac8199d1025fe1a8bafd07600eb9760762825562ecc8775dc2bbd2c81ac0
    observation: Local and linked approval prerequisites use the existing applicable resolution-review contract; done without required approval blocks, and uncertainty overrides cannot waive review.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:16Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: 30d5b2a129e15f9d0af3e7950406c93ca58fad009aa809f5ed608cc3076ca3d9
    previous: d21d9b18e28bbf29dd0ec470d0a4ee7f1fe184fee5241f649e027f99fcc6b100
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    kind: updated
    entry: /components/projects.md
    entry_revision: 470863b1c1b339088f86aad83b84f26c9d8dc92397002f09aa70f8eab52e460e
    observation: Resolver clears stale external maps, reloads linked checkouts, validates permanent project identity and records malformed/unavailable source uncertainty without remote fetching.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:18Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: 5e995e2816f06297e8a1e3974ec5a4735edb0e7db4b523b0f1451e30977d3e77
    previous: 30d5b2a129e15f9d0af3e7950406c93ca58fad009aa809f5ed608cc3076ca3d9
    contract: fee211455b12411b5720a631a7215897e33e3984f07cabbe4413e69a829e8ebf
    kind: updated
    entry: /components/cli.md
    entry_revision: e482ec704083fac87fc5fb67dd2c4c30b77d06518bb61db8be5f70f5a9e9c903
    observation: dependency show exposes current and retired override data; override requires explicit human identity/reason and current source/local snapshot. No human override was fabricated in this work.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:19Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
generation: 1
outcome: done
by: codex
at: 2026-10-15T07:22:23Z
evidence: 1bd7519
documentation:
  entries:
    - /features/cross-project.md
    - /features/lifecycle.md
    - /features/resolution-review.md
    - /components/projects.md
    - /components/cli.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---

Implemented, integrated and verified. just ci and storage race checks pass. Human resolution approval remains a separate decision.
