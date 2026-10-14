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
by: codex
at: 2026-10-14T17:45:01Z
baseline: 20261014-120009
---
