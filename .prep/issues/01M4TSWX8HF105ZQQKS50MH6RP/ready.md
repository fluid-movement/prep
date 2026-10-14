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
by: codex
at: 2026-10-13T14:33:33Z
baseline: 20261013-143332
---

User authorized proceeding with implementation; this is workflow sign-off, not manufactured human verification or exception approval.
