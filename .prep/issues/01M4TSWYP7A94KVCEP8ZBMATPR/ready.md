---
contract:
  id: 50b77acedaa1b202e7bea5a8c3a297c2a6e3933cd1230391a7f86b78350f45fc
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
    - id: 59fc51c3627aa9ceb26b3080bb41d9b9437334e4551ea67acd59d2374cadd9f6
      kind: acceptance
      text: New lifecycle, evidence, knowledge-review and approval commands provide bounded actionable guidance and consistent machine-readable success/error fields, using shared domain decisions.
    - id: 000e64a021f54eae15c92ecc4301d7f781e3222034f870acf4ebcfc68462b20f
      kind: acceptance
      text: Contract failures identify the unmet requirement/reconciliation step and a usable next command without implying that prep executes checks or authenticates actor identities.
    - id: b6cf1baf4467940c1d5d7e11b8abdca1364f39ad9f8e6578508941f4969f2d71
      kind: acceptance
      text: External workflow artifacts can be referenced with source identity/location and displayed as references; authoritative external content is not copied or treated as applicable verification solely because a link exists.
    - id: 6d9794996857024658eb8902f8ae3fb04f5fe0e49ae7555b4bc6d57be437cfe2
      kind: acceptance
      text: Existing CLI clients and legacy records remain compatible, with regression coverage for affected JSON/error contracts and external-reference round trips.
    - id: bb5e59a5ec1a103403798257bd56f47a7418766fd3716ea90e4d2fb761ae9c4c
      kind: acceptance
      text: General list/check/help output caps and guide section selection remain with their existing issues; this child verifies integration rather than implementing duplicate features.
    - id: 0c66af1e3fd19a5c2beb56cdb99b76c7426cbe524257ac917a68bb1aea2b27c0
      kind: acceptance
      text: Record focused contract-compliance checks and observed friction; no MCP server, dedicated external adapter, bidirectional sync or unmeasured small-change shortcut is introduced.
by: codex
at: 2026-10-15T07:44:58Z
baseline: 20261014-120011
---
