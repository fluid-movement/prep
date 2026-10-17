---
contract:
  id: ec949789fb403f3b823130103ccf0e2ca02c56e564374a4d5de2f8598d58d547
  verification: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
  documentation: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
  ancestors:
    - 01M52ZNHXFXHNDATT9WWQ3C87E
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
    - id: b8440e17c4f9ce41a8aa04cf5915f58e72b1744501a77d2293473170b7818fe9
      kind: acceptance
      text: prep help lists no theme command
    - id: f3b1663edb07d045a33824ea4a128cf14726d8108fd1ffa89a3345bce561815a
      kind: acceptance
      text: 'internal/palette defines exactly two palettes, and prep check rejects a config with a themes key or theme: nord'
    - id: f4979d72a82e2909484427c5935c0d589e99b6c9c6414680477c0512f010c97c
      kind: acceptance
      text: The settings screen switches between the two themes without a preview; goldens are updated, and the preview golden and six theme gallery goldens are deleted
    - id: 49958f562ae541ea3d68452a8d8af74a2498122426e131e927f445005cae3a59
      kind: acceptance
      text: Knowledge describes two themes and no custom ones, and prep check reports no errors
    - id: 5365b1495652074233b4addd8a780e10b14221009403fa96a82cb159c289be9a
      kind: acceptance
      text: go test and fmt pass, and GitHub CI on main is green
by: claude-code/opus-5.5
at: 2026-10-17T08:05:27Z
baseline: 20261017-080518
---
