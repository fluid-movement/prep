---
contract:
  id: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
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
    - id: 97a8040d8d311a503eb75680749ccf84be4c02e8799b64d9363e3bd7ad6e0a24
      kind: acceptance
      text: Findings specify a reproducible comparison of code-only, current-knowledge and lean-knowledge conditions on identical repository snapshots and tasks, including deliberately stale knowledge.
    - id: 8edbda07c94f7490ac4160fbaa57d2f4a47e6f0448d81be597735ec0bb295b2b
      kind: acceptance
      text: The comparison reuses existing evaluation infrastructure and plugin A/B work, records model/version, prompts, repetitions and retrieval conditions, and separates contract compliance from exploratory usefulness metrics.
    - id: 833c3c6dbba23fb815736a8f4ad0e070df3a554a9a113437f0b8e1a1531e4124
      kind: acceptance
      text: Report correctness, missed constraints and time/token/maintenance cost with limitations and sample sizes; do not claim superiority or add hard CI thresholds unsupported by reliable deterministic evidence.
    - id: 3b0879f79ca61f0802427e3b748dba54e660947e1ce3fc751df370fb72ab57e4
      kind: acceptance
      text: Lean examples retain decisions, constraints and navigation hints, with small briefings and targeted retrieval under the current OK Format storage contract.
    - id: 5c204fc6a924657f6f7ffb0a60a67c29104c1e90f0a84dedd2835bf80f230d08
      kind: acceptance
      text: Findings and a recommendation identify justified incremental knowledge changes; affected entries touched by this work are updated or reviewed rather than deferred.
    - id: f4f691d9eded2c3f7139547939887b0346949f7dd54c4b720f69a6b48485a72e
      kind: acceptance
      text: No lifecycle gates, alternate knowledge format or duplicate evaluation framework are implemented under this research issue.
by: codex
at: 2026-10-15T07:46:30Z
baseline: 20261014-120010
---
