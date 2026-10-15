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
verification:
  - id: a477c8a090b021155b479c33301053f5393ae17814645784caed263145c4a065
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:57.270454Z
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
  - id: ee2e97b2098d3240defdb93adca8cb461d9532b3b222c4c4937a954e633d27e4
    previous: a477c8a090b021155b479c33301053f5393ae17814645784caed263145c4a065
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Reviewed and updated affected knowledge through prep, README and synchronized harness skills. Scoped knowledge assessments record changed behavior and unchanged boundaries.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:58.676476Z
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
  - id: aba7afcc075194adf949535a3287e9cb78861ca0f9d6ba0331bc778c543bb164
    previous: ee2e97b2098d3240defdb93adca8cb461d9532b3b222c4c4937a954e633d27e4
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:00.120381Z
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
  - id: c797f428c518e186fc5a3d1848215e007d1ba6c50612eb64d0d696abce7ced07
    previous: aba7afcc075194adf949535a3287e9cb78861ca0f9d6ba0331bc778c543bb164
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:01.528437Z
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
  - id: dfe8a979ac097b16fea5920b4466a18a40020dbe6c0545d5b0ffdea0bde58fe9
    previous: c797f428c518e186fc5a3d1848215e007d1ba6c50612eb64d0d696abce7ced07
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:02.966139Z
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
  - id: 8e99ba96f83cb4faaaa2eb4d9e031a539ebdd717d3f9d754abaa86f5e4430f60
    previous: dfe8a979ac097b16fea5920b4466a18a40020dbe6c0545d5b0ffdea0bde58fe9
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Reviewed and updated affected knowledge through prep, README and synchronized harness skills. Scoped knowledge assessments record changed behavior and unchanged boundaries.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:05.437432Z
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
  - id: d4037586ba1a274d71aab7af206746f54677aa6343b8028344f7d968834559b7
    previous: 8e99ba96f83cb4faaaa2eb4d9e031a539ebdd717d3f9d754abaa86f5e4430f60
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 66f46ddcb6b9d59f92ee540bc3dec024e091f18c522d6ee9a5d163234544f508
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:06.886322Z
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
  - id: 7d0e0dcd103db8454d28c354b8c5bf765a4e15a5fdfd87a7e4530a95c09a7326
    previous: d4037586ba1a274d71aab7af206746f54677aa6343b8028344f7d968834559b7
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 7d5bda6b0171f3241d24b51014c7db8d8cc347fe0b0f0bc7740e345e802730a4
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:08.361321Z
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
  - id: 0c71e11aea3055b371f407cccc8efe120bb2b0819be97742980830f100d5bcfa
    previous: 7d0e0dcd103db8454d28c354b8c5bf765a4e15a5fdfd87a7e4530a95c09a7326
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 8a1f3b30b65011fca32a7ebd5ec0942007cb8a02b1426525500afbbbd443cc09
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:09.869344Z
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
  - id: 6de9ab0307a799ab83624924e0794c3436b1251afc7a795f2a62d43cf89760ba
    previous: 0c71e11aea3055b371f407cccc8efe120bb2b0819be97742980830f100d5bcfa
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: c4e0016b42c8458f276c29577d62e1e97c406951391af654bc373c06987937c6
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:11.381041Z
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
  - id: cf0b758653bff322fba8768a41bf7232384752979ce6873d4b5501cb359dc238
    previous: 6de9ab0307a799ab83624924e0794c3436b1251afc7a795f2a62d43cf89760ba
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 7ab9107cf1a1a8047e85afa5ceeb6b7a1f91a4034d3e58abfe73064b28a4294c
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:15.119718Z
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
  - id: 73cee29ade2f318131982cce54254f21b9a43fb221f38e858711e9b8047f0704
    previous: cf0b758653bff322fba8768a41bf7232384752979ce6873d4b5501cb359dc238
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    requirement: 37d5583bbac1b135249ba86a9f56ea993d04ce015337c529d8a15eea94918e14
    result: pass
    method: automated
    command: just ci
    observation: Passed just ci (gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check). Feature tests cover deterministic conflicting writers, stale read sets and directory membership, interrupted journal recovery, malformed recovery inputs and record-chain integrity. Final go test -race ./internal/mdstore also passed.
    provenance: reported
    by: codex
    at: 2026-10-15T07:04:18.798561Z
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
  - id: 75f58bd5692fc77e17df2ab1c30f9d617d71b8a7ac6427c1298ed531394e409e
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    kind: updated
    entry: /features/local-write-safety.md
    entry_revision: e58a99a6a97a38a4c53723d779570d5e1380fc76df508dc94c70e0ce8a638d2e
    observation: Checked OS advisory lock, frozen read-set and directory manifests, complete journal validation and bounded process-interruption roll-forward; direct-editor/distributed/power-loss guarantees are explicitly excluded.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:20Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: 8edacadcc2bb0039ef3d63d1e4ab8cdcb54d349dbb326a308f979548bd0e49fa
    previous: 75f58bd5692fc77e17df2ab1c30f9d617d71b8a7ac6427c1298ed531394e409e
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    kind: updated
    entry: /components/markdown-store.md
    entry_revision: 167d8ddc1634b796c89024c76b8faaa6133b18b35c75e6152857b63d35c44416
    observation: Locked coherent Load/Apply, CAS conflicts, empty directory canonicalization, manifest membership checks and journal recovery match tested implementation. Maintenance paths remain per-file.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:21Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: 9c69d719d8873d2af84d637682885ea4e396b4ddf5aefcca3ab15f173699ede2
    previous: 8edacadcc2bb0039ef3d63d1e4ab8cdcb54d349dbb326a308f979548bd0e49fa
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    kind: updated
    entry: /conventions/storage-format.md
    entry_revision: d6f497ebdb6814bf8e4b74c63c5737d667b5133ecc2bd3f8158cec70e37c1bc8
    observation: Dependency override sidecar hash chain and ignored lock/journal/staged metadata are described. Restart recovery validates every entry before mutation and preserves unexpected content.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:23Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: b5273904f7979a1f0bffd017f92008b40eb0deaaa774f819c92f52d8b0836941
    previous: 9c69d719d8873d2af84d637682885ea4e396b4ddf5aefcca3ab15f173699ede2
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    kind: unchanged
    entry: /components/architecture.md
    entry_revision: f746af8e3c8a0ea2d70d69417420e310030fa58ac14c442a9192f7376e698bda
    observation: 'Domain/storage boundary remains accurate: adapter owns locking, filesystem snapshots and recovery; domain remains filesystem-independent.'
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:24Z
    scopes:
      - .claude/skills/prep/SKILL.md
      - README.md
      - cmd
      - internal
      - plugins
    snapshot: 5905f693c1ff36caa34a5bed6f99590dcf0a4c96318e191d233a7b393f10eb3d
  - id: 2edc1ddadd59225ceb75c0ad34f08ebea8eb03ff5025629483c50044c5854d84
    previous: b5273904f7979a1f0bffd017f92008b40eb0deaaa774f819c92f52d8b0836941
    contract: c1cd3d202c971843c768048ae1f63be997426a7283210b9ca7061314ed8f58f0
    kind: unchanged
    entry: /decisions/tracking.md
    entry_revision: 0dcaf07fbc76f9086676d841e69987616436ae4b6b468926bf8ca1c96df050ff
    observation: Tracking through prep and human-led orchestration remain accurate; no distributed scheduler or automatic branch merging introduced.
    provenance: reported
    by: codex
    at: 2026-10-15T07:03:25Z
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
at: 2026-10-15T07:22:26Z
evidence: 1bd7519
documentation:
  entries:
    - /features/local-write-safety.md
    - /components/markdown-store.md
    - /conventions/storage-format.md
    - /components/architecture.md
    - /decisions/tracking.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---

Implemented, integrated and verified. just ci and storage race checks pass. Human resolution approval remains a separate decision.
