---
title: Hooks in the popular harnesses make sure prep calls are made
kind: code
parent: 01M498VQD4Q85CZYK965WW3H9Q
depends_on:
  - 01M4SX94VFDXCRYWNKCJG1H01J
  - 01M4J87R8MZCWPYPQ2JBPJYC8D
tags:
  - integration
  - adoption
---

Bring the shared prep hook commands, exercised in the Codex integration, to other popular agent harnesses. Reuse the shared session-start, prompt, finish and supporting before-tool/after-tool behavior through each harness-specific event protocol; installed and removed by prep setup. Each child is one harness and records any missing event or fallback. Claude-specific model verification is tracked separately and does not block this shared foundation. Preserve the accepted tracking policy, original user scope, bounded context, record redirects, fail-open errors and activity attribution. The shared mechanics live in internal/cli/hook.go and hook_finish.go and are documented in /components/hooks.md.

## Open questions
