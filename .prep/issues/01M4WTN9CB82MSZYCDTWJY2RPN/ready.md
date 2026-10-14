---
contract:
  id: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
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
    - id: 163f554bcf624f62590ecec3835ed3a8b9822570ed416fe8ff8e147d4ab5b158
      kind: acceptance
      text: Codex session-start guidance limits PREP_ACTOR/PREP_SESSION assignments to prep CLI invocations and preserves ordinary shell command shape.
    - id: 4d467d96d45416c3ea00877c04730f48c4325d3a72a2b31cc01d0ce55bee0e27
      kind: acceptance
      text: Prep read/write activity retains the supplied Codex actor and session, with explicit --by precedence.
    - id: 29e2b456d1140ceae6ad405bbed0347c8aa0b3833dbcb0145826bc9336da5aa9
      kind: acceptance
      text: Ordinary tools are neither rewritten nor auto-approved; tool events retain Codex attribution without command env assignments.
    - id: 1deeb5dc057082639be839074b359433c1493b3e58505ec2608104f4265047a7
      kind: acceptance
      text: Embedded and packaged skills, README and affected knowledge entries agree; relevant regressions and project checks pass.
by: codex
at: 2026-10-14T09:20:06Z
baseline: 20261014-091923
---
