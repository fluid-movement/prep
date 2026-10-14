---
contract:
  id: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
  verification: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
  documentation: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
  ancestors:
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
    - id: 163f554bcf624f62590ecec3835ed3a8b9822570ed416fe8ff8e147d4ab5b158
      kind: acceptance
      text: Codex session-start guidance limits PREP_ACTOR/PREP_SESSION assignments to prep CLI invocations and preserves ordinary shell command shape.
    - id: 4d467d96d45416c3ea00877c04730f48c4325d3a72a2b31cc01d0ce55bee0e27
      kind: acceptance
      text: Prep read/write activity retains the supplied Codex actor and session, with explicit --by precedence.
    - id: 29e2b456d1140ceae6ad405bbed0347c8aa0b3833dbcb0145826bc9336da5aa9
      kind: acceptance
      text: Ordinary tools are neither rewritten nor auto-approved; tool events retain Codex attribution without command env assignments.
    - id: 1deeb5dc057082639be839074b359433c1493b3e58505ec2608104f4265047a7
      kind: acceptance
      text: Embedded and packaged skills, README and affected knowledge entries agree; relevant regressions and project checks pass.
verification:
  - id: bc6a82217e7c995d775722efdfab10cb45e343b81b49cc7fa7a9ca7369c31708
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: 16504e24a480c9ff88b552d230e643350e80be39466f48ebef731ed9ef56e576
    result: pass
    method: automated
    command: just ci
    observation: 'just ci passed: gofmt clean, go vet clean, full Go suite passed, source prep check reports zero errors, and prep fmt --check passed.'
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:49.963991Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 944396d62dce33734c94abe293e7b5b64a684b747b9d6b5bff86cc99ea19bdd1
    previous: bc6a82217e7c995d775722efdfab10cb45e343b81b49cc7fa7a9ca7369c31708
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: bde616a984c0ff308abbd1915367e3c2e6369d5e5b0fd3eafe4c6539b28fa80c
    result: pass
    method: manual
    observation: Reviewed changed session guidance and attribution code; updated /components/codex.md, /components/hooks.md and /features/agent-activity.md to describe prep-only env assignments and event-payload tool attribution. README and all skill copies agree. Shortened the domain policy pointer to keep its entry within the size budget without dropping requirements.
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:50.146497Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 58c03135bd70fe1b0c3e8666010218aad175fae0bd68174d8237c885d99869cc
    previous: 944396d62dce33734c94abe293e7b5b64a684b747b9d6b5bff86cc99ea19bdd1
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: e31c092b698de0bf792de59832bf2d5a788432a41c16948143cdaaf846d4a51a
    result: pass
    method: automated
    command: go test ./...
    observation: The full Go suite passed as part of just ci.
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:50.370589Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: c480c3d71c98971cf08b151eec134d5c3655d0a4319b212432d15f66b4dc1810
    previous: 58c03135bd70fe1b0c3e8666010218aad175fae0bd68174d8237c885d99869cc
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: a7ece112258bfd7afcc4f379d160b7e20f14d6dbc49a37cfbdd332a15c9f40e3
    result: pass
    method: automated
    command: /private/tmp/prep-lifecycle check --json
    observation: Source CLI reports 0 errors; 6 warnings concern existing knowledge freshness.
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:50.598675Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 4b9ad10b01025712d75f0fa0f246af56981cd0f146fa2dcb6a0f6e0f57fea8b9
    previous: c480c3d71c98971cf08b151eec134d5c3655d0a4319b212432d15f66b4dc1810
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: cf65d6a35e0698feb5c7aeb24f45cbd24f7cce7ce3699aee1de855dfdf1bd48a
    result: pass
    method: automated
    command: /private/tmp/prep-lifecycle fmt --check
    observation: Record formatting check passed with no rewrites.
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:50.815602Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: a73e8a9364294b0577ad7456cca428cb0157b5785af1e577ca78c9722c526949
    previous: 4b9ad10b01025712d75f0fa0f246af56981cd0f146fa2dcb6a0f6e0f57fea8b9
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: 22dc9e5bcf626337b08cca1c9dec86179f52eddc3d10380a4aa47200ebf2b5b1
    result: pass
    method: manual
    observation: Reviewed changed session guidance and attribution code; updated /components/codex.md, /components/hooks.md and /features/agent-activity.md to describe prep-only env assignments and event-payload tool attribution. README and all skill copies agree. Shortened the domain policy pointer to keep its entry within the size budget without dropping requirements.
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:51.041978Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 00c88bc82a7aa21dd1677c6d3f325e2a06cd2c43a611a7ce67eae266bfe14566
    previous: a73e8a9364294b0577ad7456cca428cb0157b5785af1e577ca78c9722c526949
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: 163f554bcf624f62590ecec3835ed3a8b9822570ed416fe8ff8e147d4ab5b158
    result: pass
    method: automated
    command: go test ./internal/cli -run 'TestCodexSessionAttribution|TestPrepInvocationAndOrdinary|TestCodexSkillUses|TestHookRecordRedirects|TestHookSessionAndActivity|TestClaudeCodePlugin|TestSkill'
    observation: 'Attribution regressions passed: prep-only guidance with quoted/empty session IDs, attributed prep reads/writes and explicit --by precedence, ordinary tool events without env prefixes, unchanged before-tool permissions, and synchronized embedded/plugin skills plus generator.'
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:51.278962Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 8b2de6c797ded88bde7b1e1274a711a9f7839eae04c0cdc4db1de58b304ab81e
    previous: 00c88bc82a7aa21dd1677c6d3f325e2a06cd2c43a611a7ce67eae266bfe14566
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: 4d467d96d45416c3ea00877c04730f48c4325d3a72a2b31cc01d0ce55bee0e27
    result: pass
    method: automated
    command: go test ./internal/cli -run 'TestCodexSessionAttribution|TestPrepInvocationAndOrdinary|TestCodexSkillUses|TestHookRecordRedirects|TestHookSessionAndActivity|TestClaudeCodePlugin|TestSkill'
    observation: 'Attribution regressions passed: prep-only guidance with quoted/empty session IDs, attributed prep reads/writes and explicit --by precedence, ordinary tool events without env prefixes, unchanged before-tool permissions, and synchronized embedded/plugin skills plus generator.'
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:51.559066Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 619a43e2186879d04d355a172b46037371fe3c6f882376e67fc3e1dcc609ae27
    previous: 8b2de6c797ded88bde7b1e1274a711a9f7839eae04c0cdc4db1de58b304ab81e
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: 29e2b456d1140ceae6ad405bbed0347c8aa0b3833dbcb0145826bc9336da5aa9
    result: pass
    method: automated
    command: go test ./internal/cli -run 'TestCodexSessionAttribution|TestPrepInvocationAndOrdinary|TestCodexSkillUses|TestHookRecordRedirects|TestHookSessionAndActivity|TestClaudeCodePlugin|TestSkill'
    observation: 'Attribution regressions passed: prep-only guidance with quoted/empty session IDs, attributed prep reads/writes and explicit --by precedence, ordinary tool events without env prefixes, unchanged before-tool permissions, and synchronized embedded/plugin skills plus generator.'
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:51.802842Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
  - id: 47f5df08ee507ed2f27bfdd47b8f02ad35569db99d95c4343765c4989bd49231
    previous: 619a43e2186879d04d355a172b46037371fe3c6f882376e67fc3e1dcc609ae27
    contract: 782dd9ee85bf40b385bb8fc5b98e0541a8197d20d7b0b73ceb7265df4f840b7a
    requirement: 1deeb5dc057082639be839074b359433c1493b3e58505ec2608104f4265047a7
    result: pass
    method: automated
    command: go test ./internal/cli -run 'TestCodexSessionAttribution|TestPrepInvocationAndOrdinary|TestCodexSkillUses|TestHookRecordRedirects|TestHookSessionAndActivity|TestClaudeCodePlugin|TestSkill'
    observation: 'Attribution regressions passed: prep-only guidance with quoted/empty session IDs, attributed prep reads/writes and explicit --by precedence, ordinary tool events without env prefixes, unchanged before-tool permissions, and synchronized embedded/plugin skills plus generator.'
    provenance: reported
    by: codex
    at: 2026-10-14T09:50:52.044387Z
    scopes:
      - internal/cli
      - plugins/codex
      - plugins/claude-code/skills/prep
      - .claude/skills/prep
      - README.md
      - justfile
      - .prep/knowledge
    config_scopes:
      - go.mod
      - go.sum
      - .github/workflows/ci.yml
      - justfile
    snapshot: 4b5f3ad938b7b1fbed828df2d97fa64c2a8cd3637e4b07729bca4dae32059dc1
outcome: done
by: codex
at: 2026-10-14T09:50:52Z
documentation:
  entries:
    - /components/codex.md
    - /components/hooks.md
    - /features/agent-activity.md
    - /components/domain.md
dod:
  - go test ./... passes
  - prep check reports no errors
  - prep fmt --check passes
  - knowledge entries describing changed behavior are updated
---

Prep-only attribution guidance implemented and verified; no permission broadening or tool rewriting. Existing sessions need refreshed installation and a new session to receive new guidance.
