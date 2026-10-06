---
type: convention
title: Storage format
description: Layout of .prep, the frontmatter and body of each issue file, decision and Definition of Done syntax, canonical formatting rules.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - internal/mdstore
  - testdata/contract
confirmed_commit: e9a2ee41e40623af096046664c70fe757f1bad26
---

# Storage format

Applies when editing files in `.prep` by hand or changing the markdown adapter. Schema version 1.

- `project.md`: frontmatter `schema`; bullets under `## Definition of Done` are the project DoD.
- `config.yaml`: `commit_mode` (`off` | `all`), `views` (name → query flags).
- `issues/<YYYYMMDD-HHMMSS>/`: allowed entries are issue.md, acceptance.md, context.md, decisions.md, history.md, findings.md (research only), baselines/, ready.md, claim.md, resolution.md, attachments/. Anything else is an error.
- `issue.md`: frontmatter `title`, `kind`, `parent`, `depends_on` (in that order, nothing else). The body is the requirement; `## Open questions` holds unresolved questions.
- `acceptance.md`: `- [ ]` / `- [x]` criteria; optional `## Definition of Done` with additions and `- opt-out: <item> — <reason>`.
- `context.md`: free prose; `[text](/path.md)` links knowledge entries; backticked paths count as touched files for retrieval.
- `decisions.md`: entries `## <id>: <title>` followed by `date: YYYY-MM-DD`, optional `supersedes: <id>` and `outcome: true`, then the rationale and alternatives. Append-only. Blank lines between the heading and the metadata lines are accepted when reading; `prep fmt` removes them (canonical form: metadata directly under the heading).
- Records written only by commands: `baselines/<ts>.md` (by, at, kind, ack; body = requirement), `ready.md` (by, at, baseline), `claim.md` (by, at), `resolution.md` (outcome, by, at, reason, evidence, documentation.entries / no_impact, dod, dod_opt_outs).
- Canonical form: LF, no trailing whitespace, single blank lines outside code fences, frontmatter key order fixed, one field per line, checkboxes as `- [x]`. `prep fmt` normalizes; CI runs `prep fmt --check`.
