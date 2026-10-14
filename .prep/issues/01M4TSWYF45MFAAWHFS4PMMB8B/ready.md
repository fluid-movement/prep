---
contract:
  id: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
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
    - id: 66f46ddcb6b9d59f92ee540bc3dec024e091f18c522d6ee9a5d163234544f508
      kind: acceptance
      text: Document and reproduce concurrent check/write/rename and interrupted multi-file failure modes with deterministic tests before claiming stronger guarantees.
    - id: 7d5bda6b0171f3241d24b51014c7db8d8cc347fe0b0f0bc7740e345e802730a4
      kind: acceptance
      text: Conflicting writes within one checkout cannot silently lose another writer's records; stale writers fail with actionable conflict information and preserve prior data.
    - id: 8a1f3b30b65011fca32a7ebd5ec0942007cb8a02b1426525500afbbbd443cc09
      kind: acceptance
      text: Multi-file lifecycle writes have an explicit bounded atomicity/recovery guarantee; injected interruption tests verify recoverability and prevent silently presenting incomplete transitions as successful.
    - id: c4e0016b42c8458f276c29577d62e1e97c406951391af654bc373c06987937c6
      kind: acceptance
      text: Approval, verification and requirement/history records retain their content-chain and revision integrity through conflicts and recovery.
    - id: 7ab9107cf1a1a8047e85afa5ceeb6b7a1f91a4034d3e58abfe73064b28a4294c
      kind: acceptance
      text: Document that claims and activity are worktree/branch local and cannot reserve work globally; parallel-agent assignments remain an orchestrator responsibility.
    - id: 37d5583bbac1b135249ba86a9f56ea993d04ce015337c529d8a15eea94918e14
      kind: acceptance
      text: Keep Markdown and domain/storage boundaries, with no distributed service, remote lock, scheduler or automatic branch merging; focused conflict/recovery and project checks pass.
by: codex
at: 2026-10-14T17:45:05Z
baseline: 20261014-120011
---
