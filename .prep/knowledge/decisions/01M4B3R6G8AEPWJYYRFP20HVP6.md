---
type: decision
title: "Ignored paths are never staged"
description: "Ignored paths are never staged"
date: "2026-10-07"
issue: 01M4AHPY6HZ0ZFMMH349V1V2TC
---

`record` filters the written paths through `git check-ignore` before staging or committing, so writing the personal `config.yaml` does not run `git add` on it.

Rationale: `git add` on an ignored path fails, so without the filter every settings save warns, and commit mode all breaks. Filtering ignored paths generally is simpler than special-casing config.yaml, and it also respects any ignore rules a user adds.

Alternatives: drop `.prep/config.yaml` by name (works, but duplicates the ignore rule in code).
