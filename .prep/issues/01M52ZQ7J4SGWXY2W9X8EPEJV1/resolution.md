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
verification:
  - id: 9776d483c4999dafb758b7def1868d83c8799abc1b81f75f319603378440f856
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: go test ./... ; prep check ; prep fmt --check
    observation: go test ./... passed locally; GitHub CI run 39015306071 on e47e90a passed (test job, 1m56s)
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:55:55.850005Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: b2bd37665d8ee0684e39b4fbd2d4f616926b77ec2f6bc48fb78ca7a7d75b097a
    previous: 9776d483c4999dafb758b7def1868d83c8799abc1b81f75f319603378440f856
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Updated architecture, cli, watch, projects, tui and domain entries; deleted the desktop entry; confirmed cli-contracts, resolution-review and storage-format, which do not describe the removed parts
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:55:57.475457Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: d0076abc5f045662abeca81765dfa37bafdb1f260456802b13cdb7de81c4a57d
    previous: b2bd37665d8ee0684e39b4fbd2d4f616926b77ec2f6bc48fb78ca7a7d75b097a
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: go test ./...
    observation: all packages ok, including the new TestRemovedEntryIsHistoryForResolvedIssues
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:00.422157Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 54cf4dc35535a386bda791cb478f6f99afc9237c57397997a7fbd165f589c581
    previous: d0076abc5f045662abeca81765dfa37bafdb1f260456802b13cdb7de81c4a57d
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: prep check
    observation: 0 errors, 1 warning (I029 on 01M4WYKSGGDM7R3FSKE8451X2X, removed by the contract issue)
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:03.375939Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 5d01fca65d8fbd26bbc70ff37d4cf92a346edbe64a38e51b5a8332e5734c9ab2
    previous: 54cf4dc35535a386bda791cb478f6f99afc9237c57397997a7fbd165f589c581
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: prep fmt --check
    observation: exit 0
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:06.310799Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 7b03220af0aa3e43a030984dd628b76dadec4361c204729ba9c95617e08300f7
    previous: 5d01fca65d8fbd26bbc70ff37d4cf92a346edbe64a38e51b5a8332e5734c9ab2
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Entries describing the desktop app, prep watch and I020 were updated or deleted
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:09.25724Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 9dffdde17a43f2d1536787d9e324aa089f33082fe3dd6f8be968602ad9259800
    previous: 7b03220af0aa3e43a030984dd628b76dadec4361c204729ba9c95617e08300f7
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: 0b86a5e1386ce5eb9b248fed66530fd520db11957593b9a04e3143290cd9e358
    result: pass
    method: automated
    command: git ls-files desktop | wc -l; ls -d desktop
    observation: 0 tracked files; directory gone
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:12.2121Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 49f9259ceb6114d4e6041d19eea45735b2301fff52b6b1df9e0bf7ce077aebf5
    previous: 9dffdde17a43f2d1536787d9e324aa089f33082fe3dd6f8be968602ad9259800
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: 6159c9cec59fd32798b600802a68967da369dea848f00c2958b4d26bf70bda72
    result: pass
    method: automated
    command: git grep -i -l desktop -- ':!.prep/issues'
    observation: no files
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:15.160598Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 84583ae825771c20b70925599bbcabcac5ad58c89052dd07945ba45c8fd90530
    previous: 49f9259ceb6114d4e6041d19eea45735b2301fff52b6b1df9e0bf7ce077aebf5
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: 289be078b8c441dce1c9cc54959976dba72b18409cd77bd2e953b0a315b258a1
    result: pass
    method: automated
    command: gh run view 39015306071
    observation: ci.yml has only the test job; run 39015306071 on e47e90a succeeded
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:18.142349Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: d8d9bbadfd21c5bc8d897ab06cd59a5dd7e4dccabd38b9f9557c804a66debad4
    previous: 84583ae825771c20b70925599bbcabcac5ad58c89052dd07945ba45c8fd90530
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: b4588735979fd4ee468d41abb112687d3bbe2f3fe3a607ca438cde68c4fbdaf4
    result: pass
    method: automated
    command: just --list | grep -ic desktop
    observation: "0"
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:21.172299Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 6f5f8b6dc090fe7ef14b2db9f49206c4b7fd306ec4fb8724e605403d7dfa42ca
    previous: d8d9bbadfd21c5bc8d897ab06cd59a5dd7e4dccabd38b9f9557c804a66debad4
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    requirement: 36997567b5ee198d73d9e11b745b57e19942c2ced9216ac68d8e085bbd68cbee
    result: pass
    method: automated
    command: prep knowledge show /components/desktop.md; prep check
    observation: entry does not exist; no knowledge entry mentions the desktop app; prep check 0 errors
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:24.849165Z
    scopes:
      - .github/workflows/ci.yml
      - justfile
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
knowledge_reviews:
  - id: 3f2d4e385f05fc42c79d15250ace8cf3e5a32449ee6b740a7e3adc71a1b18a32
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    kind: updated
    entry: /components/architecture.md
    entry_revision: 37b442e0f34955f483eca82299d62df9d57b079fc973ef51f6cc3557ec44d07b
    observation: removed the desktop/ line, its details link and prep watch from the watch line
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:35Z
    scopes:
      - .github/workflows/ci.yml
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
      - justfile
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 853839e624f9c4b35892bbce4d1a90c25f9227a32e6948c99aaeb63b51a97f7c
    previous: 3f2d4e385f05fc42c79d15250ace8cf3e5a32449ee6b740a7e3adc71a1b18a32
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    kind: updated
    entry: /components/cli.md
    entry_revision: 7dbf3ea4065f0515a844c674ef6c5675feed5ca5583507d99c508df41459e171
    observation: removed the prep watch command description
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:37Z
    scopes:
      - .github/workflows/ci.yml
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
      - justfile
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: b1f89fb7330f07305d099eb6b9171a024cb80dd11405bc2cc0bcded47d98a6db
    previous: 853839e624f9c4b35892bbce4d1a90c25f9227a32e6948c99aaeb63b51a97f7c
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    kind: updated
    entry: /components/watch.md
    entry_revision: 39e875150854e2e770ebd3ccd936c6ea2530767f97f24700048b3f18492100d2
    observation: removed prep watch and the desktop client; description now names only the TUI
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:40Z
    scopes:
      - .github/workflows/ci.yml
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
      - justfile
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: d7dea5cf04cc315fbbd2d28c894ad060d92e7e684f5fa5a5519c060b44dae70d
    previous: b1f89fb7330f07305d099eb6b9171a024cb80dd11405bc2cc0bcded47d98a6db
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    kind: updated
    entry: /components/projects.md
    entry_revision: 6912c057dfad626945f94056fd05f7ca439a26e15cca2ed3966a957849115fb8
    observation: clients line names the TUI switcher instead of the desktop app
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:44Z
    scopes:
      - .github/workflows/ci.yml
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
      - justfile
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 805d865823c9077a709f803b0b7f6d061a507a9fad50f17de51d837c1827a531
    previous: d7dea5cf04cc315fbbd2d28c894ad060d92e7e684f5fa5a5519c060b44dae70d
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    kind: updated
    entry: /components/tui.md
    entry_revision: a8e5168b2e92b71b78a6e1d026ac73c1bb30ac05e1a8c2b88612ac60280a17b3
    observation: live reload no longer says it shares the watcher with prep watch
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:47Z
    scopes:
      - .github/workflows/ci.yml
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
      - justfile
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
  - id: 77907a880b176f59844e9d1da32e75bca23a12f795374eef4deb536f0edea84f
    previous: 805d865823c9077a709f803b0b7f6d061a507a9fad50f17de51d837c1827a531
    contract: a5c105dde4ec727a9e471cd9eceb010b8a45849b60711ec82079458d2ded5840
    kind: updated
    entry: /components/domain.md
    entry_revision: 8023eef7e0b590e6c234462953400815c4dfcb5eb6b31e500f047e68a12f1c54
    observation: resolved issues' knowledge references are unchecked history; I020 gone
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T07:56:51Z
    scopes:
      - .github/workflows/ci.yml
      - internal/cli
      - internal/domain
      - internal/projects/projects.go
      - justfile
    snapshot: 612a52e7a8d3e9bf1f9fdb362462380cfc952907d0670598dbeedd53dd1fab12
outcome: done
by: claude-code/opus-5.5
at: 2026-10-17T07:56:54Z
documentation:
  entries:
    - /components/architecture.md
    - /components/cli.md
    - /components/watch.md
    - /components/projects.md
    - /components/tui.md
    - /components/domain.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---
