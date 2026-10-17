---
contract:
  id: d27dceece203ddaaddeaf057857f60f5937f192e52b8d82935fde72d597d6a76
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
    - id: ebea8314584608826fe95aff866697e8f5e1fbe76e69f5b124c563fe217594fa
      kind: acceptance
      text: prep eval mutate does not exist and prep help does not mention it
    - id: c83d6856c4373b98decfcbf064f0b6ee5282ab37769e82dd8c436c58874f12ab
      kind: acceptance
      text: No CI workflow runs evals, and eval/experiments is gone
    - id: 95cec21acfd8826e0151a4a4b6f721183d8a1a79836630cf98f979acd2aeb2c6
      kind: acceptance
      text: Knowledge describes evals as local runs without mutate, and prep check reports no errors
    - id: 5365b1495652074233b4addd8a780e10b14221009403fa96a82cb159c289be9a
      kind: acceptance
      text: go test and fmt pass, and GitHub CI on main is green
by: claude-code/opus-5.5
at: 2026-10-17T08:19:38Z
baseline: 20261017-081928
---
