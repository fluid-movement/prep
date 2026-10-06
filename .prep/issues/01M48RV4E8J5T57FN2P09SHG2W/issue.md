---
title: CI check that rejects invalid hand edits of .prep
kind: code
tags:
  - adoption
---

When an agent cannot run the binary it edits `.prep` by hand, and nothing enforces the rules the commands enforce until someone runs `prep check` locally. This repository's own CI already runs `prep check` and `prep fmt --check` (built from source); projects that adopt prep have no such guard.

Give projects a ready-made check they can adopt in one step: a GitHub Actions workflow (written by `prep setup` or documented with a copy-paste snippet) that installs the pinned release, runs `prep check` and `prep fmt --check`, and fails the pull request on errors. The research in 01M46PWX38AZN55Q6T8TE93959 names what such a check must guarantee: structural validity and canonical form, and records that only commands may write (baselines, `ready.md`, `claim.md`, `resolution.md`) being consistent with the requirement and criteria they reference, so a hand-written transition that skips a gate is caught.

## Open questions

- Which hand-written transitions can `prep check` detect today, and which need new checks (for example a `ready.md` against a baseline whose requirement differs from `issue.md`)?
