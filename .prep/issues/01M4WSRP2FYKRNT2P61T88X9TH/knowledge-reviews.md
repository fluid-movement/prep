---
reviews:
  - id: eac1f5483b4811a65d0345daaf09c66943c1daa08c38afcadaf19885627295a6
    contract: 3f23421701707c46ad81dce75fd0dc3e0f499f2dc8d11dffff6d6a8f23b13a45
    kind: updated
    entry: /features/completion-policy.md
    entry_revision: 004d58ea15d7271185c4ae14debe5722a1cf11fa14c6689ddfd8596f0d690705
    observation: Updated to document per-entry and no-impact review gates, applicability invalidation, selected resolution snapshots, reported scope limits, and legacy compatibility; checked against final domain, CLI and TUI behavior.
    provenance: reported
    by: codex
    at: 2026-10-14T12:47:36Z
    scopes:
      - internal/cli
      - internal/domain
      - internal/eval/fixture
      - internal/mdstore
      - internal/tui
    snapshot: 22f427e9a247d0c23814a133f3df7e5bb766a597b099f98551cfe78cc9f847cf
  - id: d8c53e4fa713ec8d0ef07e5c258a0467be4e4fa29523dc2587a136ce168afade
    previous: eac1f5483b4811a65d0345daaf09c66943c1daa08c38afcadaf19885627295a6
    contract: 3f23421701707c46ad81dce75fd0dc3e0f499f2dc8d11dffff6d6a8f23b13a45
    kind: updated
    entry: /conventions/storage-format.md
    entry_revision: 042a9e4ac0073e49107fcfc476b120613742ed71caa51bc468d3a233d11ac01d
    observation: Updated to specify the append-only review record, bound fields, entry content revision, selected resolution snapshot, hash chain, and absent-field legacy compatibility; checked against final markdown parser and renderer.
    provenance: reported
    by: codex
    at: 2026-10-14T12:47:39Z
    scopes:
      - internal/mdstore
    snapshot: 9945c4744bdfbf537e23cad3cb87d202b62d0980d0e5e5b7bd6c0683050252be
  - id: 86f7692328b0b26e874efe1837dbe7221819c2856a803a813b416222a53bda8c
    previous: d8c53e4fa713ec8d0ef07e5c258a0467be4e4fa29523dc2587a136ce168afade
    contract: 3f23421701707c46ad81dce75fd0dc3e0f499f2dc8d11dffff6d6a8f23b13a45
    kind: updated
    entry: /components/eval.md
    entry_revision: 9a19f356be85c929f27dd50397a7e45b1c31dbcef9aabacc6e5742b8ea96bf03
    observation: Updated to describe explicit fixture completion reports, their synthetic stable go.mod work scope, and their limited role as fixture construction rather than independent checks; verified the builder writes matching no-impact reports before completion.
    provenance: reported
    by: codex
    at: 2026-10-14T12:47:42Z
    scopes:
      - internal/eval
      - internal/eval/fixture
    snapshot: caee1db9340a4deaa3ff3af25ee0a7680a5e51e7a09869c7344989e3b891b9c0
  - id: 968a2cce2763f3e63477a71da779999c1c98d88e7c1e8c0015cf21c890bc9959
    previous: 86f7692328b0b26e874efe1837dbe7221819c2856a803a813b416222a53bda8c
    contract: 3f23421701707c46ad81dce75fd0dc3e0f499f2dc8d11dffff6d6a8f23b13a45
    kind: updated
    entry: /features/knowledge-lifecycle.md
    entry_revision: 4131a4c657c04d5055fdd049eac09c6d81c5bbdedb24f7e512e9d90564326723
    observation: Added a concise feature reference for the bounded guide candidates, report/show commands, explicit completion alignment, applicability limits, legacy boundary, and selected resolution reports; checked against final domain, CLI and TUI gates.
    provenance: reported
    by: codex
    at: 2026-10-14T12:48:30Z
    scopes:
      - internal/cli
      - internal/domain
      - internal/mdstore
      - internal/tui
    snapshot: cc408bd647947e7c1fecc8d8451af0665a7f5661e0d4c8282cc2bb459e1765d2
---
