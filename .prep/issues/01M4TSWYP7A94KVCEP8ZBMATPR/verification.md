---
evidence:
  - id: 2c5a02be352b431c59fe2368bb510cb83d402b330ee3c1d843e8082fec345b4d
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: 'Passed just ci: gofmt, go vet, go test ./... including CLI/domain/storage and TUI suites; prep check 0 errors/0 warnings and prep fmt --check. Targeted artifact and lifecycle compatibility tests also passed.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:10.135279Z
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
    snapshot: b4ac0db1453a22fd5de34eae4d63d8eb624d6ff60159daa55781dfcee8a68319
  - id: b2d0ea2f8876d9b64d8693aeb820b842e37caec606e394b9755aee6635f6a2ce
    previous: 2c5a02be352b431c59fe2368bb510cb83d402b330ee3c1d843e8082fec345b4d
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:12.066539Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: 43ae6f178987ca71e34a4d29ff84153a9cb5bc19ac9cd4f2a2f530a3c5eeb5d5
    previous: b2d0ea2f8876d9b64d8693aeb820b842e37caec606e394b9755aee6635f6a2ce
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: 'Passed just ci: gofmt, go vet, go test ./... including CLI/domain/storage and TUI suites; prep check 0 errors/0 warnings and prep fmt --check. Targeted artifact and lifecycle compatibility tests also passed.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:14.085595Z
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
    snapshot: b4ac0db1453a22fd5de34eae4d63d8eb624d6ff60159daa55781dfcee8a68319
  - id: 04f042d4518a710f5fac0b78bad1483f54144ae875fc83d50abc57f662ebb39d
    previous: 43ae6f178987ca71e34a4d29ff84153a9cb5bc19ac9cd4f2a2f530a3c5eeb5d5
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: just ci
    observation: 'Passed just ci: gofmt, go vet, go test ./... including CLI/domain/storage and TUI suites; prep check 0 errors/0 warnings and prep fmt --check. Targeted artifact and lifecycle compatibility tests also passed.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:16.609174Z
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
    snapshot: b4ac0db1453a22fd5de34eae4d63d8eb624d6ff60159daa55781dfcee8a68319
  - id: 680eca4db8ae564ed2378a5c6a315efd3471c8e1bf8d8ff1c83883d39fa01e29
    previous: 04f042d4518a710f5fac0b78bad1483f54144ae875fc83d50abc57f662ebb39d
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: just ci
    observation: 'Passed just ci: gofmt, go vet, go test ./... including CLI/domain/storage and TUI suites; prep check 0 errors/0 warnings and prep fmt --check. Targeted artifact and lifecycle compatibility tests also passed.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:18.721533Z
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
    snapshot: b4ac0db1453a22fd5de34eae4d63d8eb624d6ff60159daa55781dfcee8a68319
  - id: 6b168f7a8c07591091242e8a7aa73553985a229433b3b7dec4bb33644f452efe
    previous: 680eca4db8ae564ed2378a5c6a315efd3471c8e1bf8d8ff1c83883d39fa01e29
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: knowledge entries describing changed behavior are updated Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:22.292354Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: eb14c685ae38e27e879741d1a73172a255f5ee6402fd835fddb4c51147ad59b4
    previous: 6b168f7a8c07591091242e8a7aa73553985a229433b3b7dec4bb33644f452efe
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: 59fc51c3627aa9ceb26b3080bb41d9b9437334e4551ea67acd59d2374cadd9f6
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: New lifecycle, evidence, knowledge-review and approval commands provide bounded actionable guidance and consistent machine-readable success/error fields, using shared domain decisions. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:24.330692Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: bf58020b5a6f58858f96ed2c862c6f5bdf37ce87f75f2eb3ee23d886eef4a1ba
    previous: eb14c685ae38e27e879741d1a73172a255f5ee6402fd835fddb4c51147ad59b4
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: 000e64a021f54eae15c92ecc4301d7f781e3222034f870acf4ebcfc68462b20f
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: Contract failures identify the unmet requirement/reconciliation step and a usable next command without implying that prep executes checks or authenticates actor identities. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:26.385539Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: ebce6a4d36065a5e8ae3f6e2638a060df023d421ec4810cf574be4670c61559d
    previous: bf58020b5a6f58858f96ed2c862c6f5bdf37ce87f75f2eb3ee23d886eef4a1ba
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: b6cf1baf4467940c1d5d7e11b8abdca1364f39ad9f8e6578508941f4969f2d71
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: External workflow artifacts can be referenced with source identity/location and displayed as references; authoritative external content is not copied or treated as applicable verification solely because a link exists. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:28.398501Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: 58d51c084de160b6c5e9d7b14c5ee7819e906d0eca332f7aa0b89349f801682c
    previous: ebce6a4d36065a5e8ae3f6e2638a060df023d421ec4810cf574be4670c61559d
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: 6d9794996857024658eb8902f8ae3fb04f5fe0e49ae7555b4bc6d57be437cfe2
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: Existing CLI clients and legacy records remain compatible, with regression coverage for affected JSON/error contracts and external-reference round trips. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:30.432467Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: e0e5d0dde8d32ff6549c9ac9ad5618b230941058b644a323ed8be466dc19f2e9
    previous: 58d51c084de160b6c5e9d7b14c5ee7819e906d0eca332f7aa0b89349f801682c
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: bb5e59a5ec1a103403798257bd56f47a7418766fd3716ea90e4d2fb761ae9c4c
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: General list/check/help output caps and guide section selection remain with their existing issues; this child verifies integration rather than implementing duplicate features. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:32.501246Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
  - id: cf25639129a3760bf4e6b67d2fea2bf0618dcccb94ed8471ba28ed11fcdccd5c
    previous: e0e5d0dde8d32ff6549c9ac9ad5618b230941058b644a323ed8be466dc19f2e9
    contract: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
    requirement: 0c66af1e3fd19a5c2beb56cdb99b76c7426cbe524257ac917a68bb1aea2b27c0
    result: pass
    method: manual
    observation: 'Reviewed implementation and focused contract tests against this requirement: Record focused contract-compliance checks and observed friction; no MCP server, dedicated external adapter, bidirectional sync or unmeasured small-change shortcut is introduced. Reports retain provenance and limits; references do not satisfy gates. General caps, guide selectors, MCP, sync and dedicated adapters remain outside this child.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:01:34.599641Z
    scopes:
      - internal
      - cmd
      - README.md
      - plugins
      - .claude/skills/prep/SKILL.md
      - .prep/knowledge/features/cli-contracts.md
      - .prep/knowledge/components/cli.md
      - .prep/knowledge/components/domain.md
      - .prep/knowledge/components/markdown-store.md
      - .prep/knowledge/conventions/storage-format.md
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: 7ccfdedc1bf3efffac2ab9efdc46c85e6f1c05b822d3324a12075829057b5ddf
---
