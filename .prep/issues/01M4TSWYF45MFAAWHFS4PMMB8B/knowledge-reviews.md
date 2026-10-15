---
reviews:
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
---
