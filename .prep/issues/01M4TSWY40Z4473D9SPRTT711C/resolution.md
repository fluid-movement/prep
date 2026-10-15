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
verification:
  - id: 0348aae68ceaea91496e18a2ec3a7f34c0bbdda1472c8e9044b9773928344909
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: 'Passed final just ci: gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check. Python driver syntax, unique 12-run manifest, common overview/skill hashes and command-classifier negative controls also verified; real pilot observations and limitations are recorded separately.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:12.126819Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: d27c7092c4b9f89a1909e898fdcc31a5e40b43b88aad55df376faaa5326e1885
  - id: 952ab308e586e12390a5072c6d1ffe6911afa828b80744beb62db7eb61b762ca
    previous: 0348aae68ceaea91496e18a2ec3a7f34c0bbdda1472c8e9044b9773928344909
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: 'Assessed this requirement: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:15.254469Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: cb007affca1218af1bc55030a3c94229ac44756680ac8a484eee09e907fd3a86
    previous: 952ab308e586e12390a5072c6d1ffe6911afa828b80744beb62db7eb61b762ca
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: just ci
    observation: 'Passed final just ci: gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check. Python driver syntax, unique 12-run manifest, common overview/skill hashes and command-classifier negative controls also verified; real pilot observations and limitations are recorded separately.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:18.630528Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: d27c7092c4b9f89a1909e898fdcc31a5e40b43b88aad55df376faaa5326e1885
  - id: 52de3edbd60ce9720aaed109b73116cc21489dde7a07e8e5bd528dc7a1f330a0
    previous: cb007affca1218af1bc55030a3c94229ac44756680ac8a484eee09e907fd3a86
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: just ci
    observation: 'Passed final just ci: gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check. Python driver syntax, unique 12-run manifest, common overview/skill hashes and command-classifier negative controls also verified; real pilot observations and limitations are recorded separately.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:21.931043Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: d27c7092c4b9f89a1909e898fdcc31a5e40b43b88aad55df376faaa5326e1885
  - id: 4402e7da00f8cae3423aca124d9d0f145873fc1207097f6d373cc08a8d6b869d
    previous: 52de3edbd60ce9720aaed109b73116cc21489dde7a07e8e5bd528dc7a1f330a0
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: just ci
    observation: 'Passed final just ci: gofmt, go vet, go test ./..., prep check 0 errors/0 warnings and prep fmt --check. Python driver syntax, unique 12-run manifest, common overview/skill hashes and command-classifier negative controls also verified; real pilot observations and limitations are recorded separately.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:25.2836Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: d27c7092c4b9f89a1909e898fdcc31a5e40b43b88aad55df376faaa5326e1885
  - id: c9bbcf020c4e3327d2a482d68a502d76c8283ed87739a8dfe15980f239451e5a
    previous: 4402e7da00f8cae3423aca124d9d0f145873fc1207097f6d373cc08a8d6b869d
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: 'Assessed this requirement: knowledge entries describing changed behavior are updated Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:28.960587Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: c402c1e5a736a050bd0157f293e2001cb87e5e1ffc7eb5d090b19a4bc6903ae0
    previous: c9bbcf020c4e3327d2a482d68a502d76c8283ed87739a8dfe15980f239451e5a
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 97a8040d8d311a503eb75680749ccf84be4c02e8799b64d9363e3bd7ad6e0a24
    result: pass
    method: manual
    observation: 'Assessed this requirement: Findings specify a reproducible comparison of code-only, current-knowledge and lean-knowledge conditions on identical repository snapshots and tasks, including deliberately stale knowledge. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:32.669383Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: 3a23925eed0a55b17eac326b3356d9db2225c9af289426126f8b6908a47c3b96
    previous: c402c1e5a736a050bd0157f293e2001cb87e5e1ffc7eb5d090b19a4bc6903ae0
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 8edbda07c94f7490ac4160fbaa57d2f4a47e6f0448d81be597735ec0bb295b2b
    result: pass
    method: manual
    observation: 'Assessed this requirement: The comparison reuses existing evaluation infrastructure and plugin A/B work, records model/version, prompts, repetitions and retrieval conditions, and separates contract compliance from exploratory usefulness metrics. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:36.510005Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: ae59ee3cdcf6436a6e20c9bfa8b447999a8c036eeb09b91b4dcd504a90fa409f
    previous: 3a23925eed0a55b17eac326b3356d9db2225c9af289426126f8b6908a47c3b96
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 833c3c6dbba23fb815736a8f4ad0e070df3a554a9a113437f0b8e1a1531e4124
    result: pass
    method: manual
    observation: 'Assessed this requirement: Report correctness, missed constraints and time/token/maintenance cost with limitations and sample sizes; do not claim superiority or add hard CI thresholds unsupported by reliable deterministic evidence. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:40.52775Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: a67d2979505746c284c2c520ed879482a08ae60aef36d1921fd66eb963f59397
    previous: ae59ee3cdcf6436a6e20c9bfa8b447999a8c036eeb09b91b4dcd504a90fa409f
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 3b0879f79ca61f0802427e3b748dba54e660947e1ce3fc751df370fb72ab57e4
    result: pass
    method: manual
    observation: 'Assessed this requirement: Lean examples retain decisions, constraints and navigation hints, with small briefings and targeted retrieval under the current OK Format storage contract. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:44.47818Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: 4fe88aa441201c9d89fef4df76be2a6143fc3104c0536c1b1000e3c6885b3e44
    previous: a67d2979505746c284c2c520ed879482a08ae60aef36d1921fd66eb963f59397
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: 5c204fc6a924657f6f7ffb0a60a67c29104c1e90f0a84dedd2835bf80f230d08
    result: pass
    method: manual
    observation: 'Assessed this requirement: Findings and a recommendation identify justified incremental knowledge changes; affected entries touched by this work are updated or reviewed rather than deferred. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:48.785079Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
  - id: 7a90bff0e82c93e271b716a33271e34360bbb8c45a7bb0bab34a8146f206a566
    previous: 4fe88aa441201c9d89fef4df76be2a6143fc3104c0536c1b1000e3c6885b3e44
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    requirement: f4f691d9eded2c3f7139547939887b0346949f7dd54c4b720f69a6b48485a72e
    result: pass
    method: manual
    observation: 'Assessed this requirement: No lifecycle gates, alternate knowledge format or duplicate evaluation framework are implemented under this research issue. Research artifacts record 12 matched Codex sessions (3 per condition), exact source/task/skill manifests, model request versus unreported server model, usage/time and maintenance proxies. All 12 helper runtime checks passed; original helper negative control failed. Only one actual test.sh execution occurred and failed; overlays dropped a mandatory test convention, so no usefulness ranking or hard threshold is claimed. Mandatory constraints are retained in the recommended lean example. All scoped implementation and documentation are integrated.'
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:52.959813Z
    scopes:
      - eval/experiments/knowledge-policy
      - internal/eval
      - eval/fixtures/webapp.yaml
      - eval/cases/natural/fix-matching-issue.yaml
      - .prep/knowledge
    config_scopes:
      - justfile
      - go.mod
      - go.sum
    snapshot: f529a980c6d14ac3df3524721577c26014b54022ea912599cd4d18bdff2ca4b3
knowledge_reviews:
  - id: f4e902d06ab2915e84b647af3289c7711d5c03f3536e9f0175aebc25ba1fca95
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    kind: updated
    entry: /features/knowledge-policy.md
    entry_revision: 02ce77d17cec8d1b5fe4ad022c8a37ae2437f532224173973c067961aa6ced22
    observation: Reviewed the integrated lifecycle and bounded Codex comparison against code, fixture protocol and recorded outcomes. Lean policy retains decisions, constraints, navigation and mandatory checks; pilot limitations and semantic trust limits are explicit. Existing gates and authorization boundaries remain accurate.
    provenance: reported
    by: codex
    at: 2026-10-15T08:29:56Z
    scopes:
      - eval/cases/natural/fix-matching-issue.yaml
      - eval/experiments/knowledge-policy
      - eval/fixtures/webapp.yaml
      - internal/eval
    snapshot: 0c7eea5c3f1870c22c5ad5fe13b728dc9d3a9f37242fcf6a373620bb0fb60891
  - id: 0fc53026e771aa04ad9e37b2c8674fc8247a1a97cb244e34be93d56602070184
    previous: f4e902d06ab2915e84b647af3289c7711d5c03f3536e9f0175aebc25ba1fca95
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    kind: updated
    entry: /conventions/knowledge-base.md
    entry_revision: 131baa2cd5a32db53043a6dd9e0c253e4d35e4c579a6f28dd055c3f0ffdcf9d4
    observation: Reviewed the integrated lifecycle and bounded Codex comparison against code, fixture protocol and recorded outcomes. Lean policy retains decisions, constraints, navigation and mandatory checks; pilot limitations and semantic trust limits are explicit. Existing gates and authorization boundaries remain accurate.
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:00Z
    scopes:
      - eval/cases/natural/fix-matching-issue.yaml
      - eval/experiments/knowledge-policy
      - eval/fixtures/webapp.yaml
      - internal/eval
    snapshot: 0c7eea5c3f1870c22c5ad5fe13b728dc9d3a9f37242fcf6a373620bb0fb60891
  - id: 68faa1f3350c0b305dc64c652074aa3a381107781fc7a7e5510b48b0bdec9bfb
    previous: 0fc53026e771aa04ad9e37b2c8674fc8247a1a97cb244e34be93d56602070184
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    kind: updated
    entry: /components/eval.md
    entry_revision: 5ff8c26b0e8dcfb8aa7ce000ac33732f5ddb9fc4c34726d6a66f5b562d14b1d9
    observation: Reviewed the integrated lifecycle and bounded Codex comparison against code, fixture protocol and recorded outcomes. Lean policy retains decisions, constraints, navigation and mandatory checks; pilot limitations and semantic trust limits are explicit. Existing gates and authorization boundaries remain accurate.
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:03Z
    scopes:
      - eval/cases/natural/fix-matching-issue.yaml
      - eval/experiments/knowledge-policy
      - eval/fixtures/webapp.yaml
      - internal/eval
    snapshot: 0c7eea5c3f1870c22c5ad5fe13b728dc9d3a9f37242fcf6a373620bb0fb60891
  - id: e534ffc7a6d5abe247fb132dd81da079dc6f4792dd592badaabbf695f851614f
    previous: 68faa1f3350c0b305dc64c652074aa3a381107781fc7a7e5510b48b0bdec9bfb
    contract: 566c07e28a8413cca70e19f8a72376afc2ff2c328718841a617b1f6e6241c4d4
    kind: unchanged
    entry: /components/eval-cases.md
    entry_revision: 5343765b39e6a3e7e388571adc37b6c1a96e4fe31524ba28fb79116dd6ad5976
    observation: Reviewed the integrated lifecycle and bounded Codex comparison against code, fixture protocol and recorded outcomes. Lean policy retains decisions, constraints, navigation and mandatory checks; pilot limitations and semantic trust limits are explicit. Existing gates and authorization boundaries remain accurate.
    provenance: reported
    by: codex
    at: 2026-10-15T08:30:06Z
    scopes:
      - eval/cases/natural/fix-matching-issue.yaml
      - eval/experiments/knowledge-policy
      - eval/fixtures/webapp.yaml
      - internal/eval
    snapshot: 0c7eea5c3f1870c22c5ad5fe13b728dc9d3a9f37242fcf6a373620bb0fb60891
generation: 1
outcome: done
by: codex
at: 2026-10-15T08:31:02Z
documentation:
  entries:
    - /features/knowledge-policy.md
    - /conventions/knowledge-base.md
    - /components/eval.md
    - /components/eval-cases.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---

Bounded comparison completed with honest negative findings and limitations; lean constraint-retention policy documented. No statistical superiority or semantic-proof claim.
