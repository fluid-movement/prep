---
contract:
  id: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
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
    - id: 805acd5942fdc00912236e604b1bfeb8594471165e51b712ce859e9a85e33f60
      kind: acceptance
      text: An entry whose root directory was deleted no longer appears in prep projects or the briefing, and the registry file no longer holds it
    - id: e66696dfa6e9b2dac9d06562dbb82cfec91cb64626f0e116976f743a6861340e
      kind: acceptance
      text: Existing projects, --project resolution by name, ID, remote or path, and prep projects add keep working; tests cover a deleted root
    - id: ad8f961c41b46c677c5534dd94f627f151fed29c23579beb930163fac8f17dc5
      kind: acceptance
      text: The Project registry knowledge entry describes the rule, and prep check reports no errors
    - id: 5365b1495652074233b4addd8a780e10b14221009403fa96a82cb159c289be9a
      kind: acceptance
      text: go test and fmt pass, and GitHub CI on main is green
by: claude-code/opus-5.5
at: 2026-10-17T08:20:05Z
baseline: 20261017-081949
---
