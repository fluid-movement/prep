---
contract:
  id: 2dc3202504570545e259c64ee49399320f682b1d734d2ed0c669821ffca05337
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
    - id: a56be183f6b81168e673ad5847d6d693acc15bbafc03ce744c95194f1cb16dd9
      kind: acceptance
      text: prep help lists no review or dependency override command, and no record in this repository has review, approval-prerequisite or override files or fields
    - id: 43e94dd242ec10d36fe9e06b3d617285b6a0ef743bb7cf95b5fa7d66b163b26d
      kind: acceptance
      text: The TUI's screens are Issues, Agent and Knowledge; goldens are updated and no Inbox code remains
    - id: f2249e81f343568bf3ed463351a3348eb0ea024c3de17854b46fbdbb5cef2a32
      kind: acceptance
      text: Reopening a done issue with a reason appends its resolution to history.md and leaves it ready, covered by a test
    - id: e4aecae75bfa3caa38a4b94e548ed21dcb721127f4ddc1f07b8bfcc9c0f28668
      kind: acceptance
      text: git grep finds no resolution review, generation, approval prerequisite or Inbox types
    - id: 7ef3a5405f485ce14e3ea7cd9cdbefd1421cccca1fb8dd554a55de3800311ccd
      kind: acceptance
      text: Knowledge describing review, the Inbox and reopen is deleted or updated, prep check reports no errors, and GitHub CI on main is green
by: claude-code/opus-5.5
at: 2026-10-17T08:20:28Z
baseline: 20261017-082017
---
