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
verification:
  - id: af1a56d7de4d6650148b936ca938bcb91bd11c8f45e118bea3b74762df16362d
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: go test ./internal/projects ./internal/cli; prep check; prep fmt --check
    observation: projects and cli tests pass; GitHub CI run 39018683224 on 78e4666 succeeded
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:49:47.779086Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 759fdf1752440bbb5e266feb9da71cf2d0cd7a400446d65429ee11ac491dbb4c
    previous: af1a56d7de4d6650148b936ca938bcb91bd11c8f45e118bea3b74762df16362d
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Project registry entry describes pruning; architecture, cli, cli-contracts and resolution-review drifted only through a test file and were confirmed
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:49:49.810132Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 58a02e0f2e58c48e7ab5a03fc6a8bba2033fd3b489242350134ed491a63aeb4e
    previous: 759fdf1752440bbb5e266feb9da71cf2d0cd7a400446d65429ee11ac491dbb4c
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: go test ./...
    observation: GitHub CI run 39018683224 ran go test ./... and succeeded
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:49:54.72542Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 000343b7afa9bb69f232f3a28267707b1f1ac26a16fe8eb1d65dfff9c2473806
    previous: 58a02e0f2e58c48e7ab5a03fc6a8bba2033fd3b489242350134ed491a63aeb4e
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: prep check
    observation: 0 errors, 1 warning (I029)
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:50:02.620938Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 7fb17df494a1d61da62691869cdd373fcdaed380a1f4247c1ac079408c17df3b
    previous: 000343b7afa9bb69f232f3a28267707b1f1ac26a16fe8eb1d65dfff9c2473806
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: prep fmt --check
    observation: exit 0
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:50:17.93523Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 9e924c8ab534abb75df3c377a436b35a8cc66f93f695d5d69c7861e864f9ced0
    previous: 7fb17df494a1d61da62691869cdd373fcdaed380a1f4247c1ac079408c17df3b
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Project registry entry updated
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:50:24.494933Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: f8c2fa7a473b0eb02b94fe4b2028b6568e7f6f0bfde10b317153a2c6b6f4cc3a
    previous: 9e924c8ab534abb75df3c377a436b35a8cc66f93f695d5d69c7861e864f9ced0
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: 805acd5942fdc00912236e604b1bfeb8594471165e51b712ce859e9a85e33f60
    result: pass
    method: automated
    command: go test ./internal/projects
    observation: load drops an entry whose root was deleted and saves the cleaned registry; List and the briefing no longer show it
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:50:40.692231Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 39cb38056a17a20ba44f1d921b6e32e077047c6d524b235747086f9a0441d632
    previous: f8c2fa7a473b0eb02b94fe4b2028b6568e7f6f0bfde10b317153a2c6b6f4cc3a
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: e66696dfa6e9b2dac9d06562dbb82cfec91cb64626f0e116976f743a6861340e
    result: pass
    method: automated
    command: go test ./internal/projects ./internal/cli -run Project
    observation: existing projects, --project resolution and projects add keep working; tests cover a deleted root
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:50:57.309823Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 3f504659505f1b69e1c763d297f44bf2d475f32366172fa0faa62728bdcdbda5
    previous: 39cb38056a17a20ba44f1d921b6e32e077047c6d524b235747086f9a0441d632
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: ad8f961c41b46c677c5534dd94f627f151fed29c23579beb930163fac8f17dc5
    result: pass
    method: automated
    command: prep knowledge show /components/projects.md; prep check
    observation: entry states that every load drops entries whose root no longer exists; 0 errors
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:51:09.365537Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
  - id: 2f56e9f5a9a4ffaaf3a0b939f8ed816c8d3cfa971ddc8bfe4bfe96d2b71d540d
    previous: 3f504659505f1b69e1c763d297f44bf2d475f32366172fa0faa62728bdcdbda5
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    requirement: 5365b1495652074233b4addd8a780e10b14221009403fa96a82cb159c289be9a
    result: pass
    method: automated
    command: gh run view 39018683224
    observation: go test and fmt pass; GitHub CI on 78e4666 succeeded
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:51:18.987207Z
    scopes:
      - internal/projects
      - internal/cli/projects_test.go
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
knowledge_reviews:
  - id: 1f3931201fc3e57cdd57bf094b50c82efd1091216f6cd608146185d3312ca85a
    contract: d2dd28111eaa3421b60e6ce5417a6b4e8ed88ca12dd001de06b3754023272743
    kind: updated
    entry: /components/projects.md
    entry_revision: ed462a4ac893e68655cce258e5c1418f5cf329e64d1db62e8d6d5f0f987e54cd
    observation: registry loads drop entries whose root no longer exists
    provenance: reported
    by: claude-code/opus-5.5
    at: 2026-10-17T08:51:26Z
    scopes:
      - internal/cli/projects_test.go
      - internal/projects
    snapshot: 0b274428587213efcf5811c3673bf65cfc3813d8e3916ffa4d87c3e4a97dcb3f
outcome: done
by: claude-code/opus-5.5
at: 2026-10-17T08:51:38Z
documentation:
  entries:
    - /components/projects.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---
