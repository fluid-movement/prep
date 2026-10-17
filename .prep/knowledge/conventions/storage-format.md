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
confirmed_commit: f5d307973808fcad23643143f1f2243ac9d54469
---

# Storage format

Applies when editing files in `.prep` by hand or changing the markdown adapter. Schema version 1; a project written by a newer schema is refused with E_SCHEMA until prep is updated.

- `project.md`: frontmatter `schema`, `id` (the project’s permanent ID, a ULID written once by `prep init` and shared by every clone; never edited by hand) and `name` (its short handle: lowercase letters, digits, `-`, `_`, `.`; `prep init` derives it from origin’s repository name, else the directory), in that order; a missing `id` is P008 (warning, `prep fix` writes one), an invalid `id` or `name` P009. Bullets under `## Definition of Done` are the project DoD.
- `.prep/local/` holds per-machine state, ignored by `.prep/.gitignore` (`local/`): the activity stream `activity.jsonl` ([Agent activity](/features/agent-activity.md)).
- `config.yaml.dist` (committed, the project’s default) and `config.yaml` (each user’s own, ignored by `.prep/.gitignore`): the own file replaces the dist file as a whole when it exists. Both hold leading `#` comment lines, `views` (name → query flags, in tab order) and `theme` (one of the built-in theme names; see [Palette and themes](/components/palette.md)). The TUI settings write the own file in this form, creating it from the dist file and keeping the leading comments; other comments are not kept. Unknown keys are errors (P003).
- `issues/<id>/`: ULID-named issue directories hold issue.md, acceptance.md, context.md, history.md, baselines/, ready.md, claim.md, resolution.md and attachments/. New issues contain only issue.md; optional empty record files are absent. Legacy decisions.md and findings.md are tolerated with K008 warnings until converted to linked knowledge records.
- `issue.md`: frontmatter title, kind, parent, depends_on, tags, priority; body is the requirement and Open questions section. Duplicate headings outside fences in issue.md, acceptance.md or context.md are I028; repeated Open questions is an error, other repeats warnings.
- `acceptance.md`: `- [ ]` / `- [x]` criteria only. The Definition of Done is project-level guidance: bullets under `## Definition of Done` in `project.md`, edited with `prep dod --add`/`--remove`.
- `context.md`: free prose; `[text](/path.md)` links knowledge entries; backticked paths count as touched files for retrieval.
- `knowledge/decisions/<ulid>.md` and `knowledge/findings/<ulid>.md`: immutable OKF records with required type, title, one-line description and date; optional issue, scope globs, decision supersedes record IDs and outcome: true; body holds rationale or findings. The filename supplies the ID. Superseded status is derived from newer records, never written onto older ones. Indexes are generated.
- Records written only by commands: baselines/<timestamp>.md (by, at, kind, optional ack, requirement body), ready.md (by, at, baseline), claim.md (by, at), resolution.md (outcome, by, at, optional reason/evidence/note). Legacy resolution documentation fields remain readable.
- Canonical form: LF, no trailing whitespace, single blank lines outside code fences, frontmatter key order fixed, one field per line, checkboxes as `- [x]`. `prep fmt` normalizes; CI runs `prep fmt --check`.
- Repository .gitattributes carries `.prep/**/history.md merge=union` and `.prep/knowledge/**/index.md merge=union`; init appends missing rules. Run prep fmt or prep fix after a union merge to regenerate indexes; no custom merge driver is needed.
- Local storage metadata: `.prep/local/write.lock` serializes cooperating readers/writers, and sibling `.prep-tmp-*` files are the temporary files of atomic writes. These are not domain records. See [Local write safety](/features/local-write-safety.md) for the bounded guarantee.
