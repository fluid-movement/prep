---
contract:
  id: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
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
    - id: 0b86a5e1386ce5eb9b248fed66530fd520db11957593b9a04e3143290cd9e358
      kind: acceptance
      text: git ls-files desktop is empty and the local build output is gone
    - id: 6159c9cec59fd32798b600802a68967da369dea848f00c2958b4d26bf70bda72
      kind: acceptance
      text: git grep -i desktop outside issue records finds nothing
    - id: 289be078b8c441dce1c9cc54959976dba72b18409cd77bd2e953b0a315b258a1
      kind: acceptance
      text: CI has no desktop job and GitHub CI on main is green after the push
    - id: b4588735979fd4ee468d41abb112687d3bbe2f3fe3a607ca438cde68c4fbdaf4
      kind: acceptance
      text: just --list shows no desktop recipes
    - id: 36997567b5ee198d73d9e11b745b57e19942c2ced9216ac68d8e085bbd68cbee
      kind: acceptance
      text: The desktop knowledge entry is deleted, no other entry mentions the app, and prep check reports no errors
by: claude-code/opus-5.5
at: 2026-10-17T07:39:28Z
baseline: 20261017-073919
---
