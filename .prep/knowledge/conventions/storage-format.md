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
confirmed_commit: 8030cc0c6ca5ae85dd8d4b1e186d0aab1b963960
---

# Storage format

Applies when editing files in `.prep` by hand or changing the markdown adapter. Schema version 1.

- `project.md`: frontmatter `schema`, `id` (the project's permanent ID, a ULID written once by `prep init` and shared by every clone; never edited by hand) and `name` (its short handle: lowercase letters, digits, `-`, `_`, `.`; `prep init` derives it from origin's repository name, else the directory), then `links` (local name → `id`, optional `remote`, names sorted; see [Cross-project work](/features/cross-project.md)), in that order; a missing `id` is P008 (warning, `prep fix` writes one), an invalid `id` or `name` P009. Bullets under `## Definition of Done` are the project DoD.
- `.prep/local/` holds per-machine state, ignored by `.prep/.gitignore` (`local/`): the activity stream `activity.jsonl` ([Agent activity](/features/agent-activity.md)).
- `config.yaml.dist` (committed, the project's default) and `config.yaml` (each user's own, ignored by `.prep/.gitignore`): the own file replaces the dist file as a whole when it exists. Both hold leading `#` comment lines, `views` (name → query flags, in tab order), `theme` (a theme name) and `themes` (custom themes: `base` and token → `"#RRGGBB"` pairs under the theme name; see [Palette and themes](/components/palette.md)). The TUI settings write the own file in this form, creating it from the dist file, keeping the leading comments; other comments are not kept. Unknown keys are errors (P003); the removed `commit_mode` is the exception, read as the prep 0.1.0 layout and reported as P006, which `prep fix` converts.
- `issues/<id>/`: the ID is a ULID (26 characters of Crockford base32, upper case: 48 bits of creation milliseconds and 80 random bits), so IDs created on different branches do not collide and sort by creation time as strings. Any other directory name is I001. Allowed entries are issue.md, acceptance.md, context.md, decisions.md, history.md, findings.md (research only), baselines/, ready.md, claim.md, verification.md, approvals.md, knowledge-reviews.md, resolution-history.md, resolution-reviews.md, dependency-overrides.md, resolution.md, attachments/. Anything else is an error.
- `issue.md`: frontmatter `title`, `kind`, `parent`, `depends_on`, optional `depends_on_approval`, `tags`, `priority` (in that order, nothing else); tags are lowercase letters, digits and `. _ - /`; priority is `critical`, `high` or `low`, absent meaning medium (medium is never written; added within schema 1). The body is the requirement; `## Open questions` holds unresolved questions. A `##` heading appears at most once per record file (`issue.md`, `acceptance.md`, `context.md`, `findings.md`; headings in code fences do not count): prep check reports a repeat as I028, an error for Open questions, which prep reads only the first copy of, and a warning otherwise.

- `acceptance.md`: `- [ ]` / `- [x]` criteria; optional `## Definition of Done` with additions and `- opt-out: <item> — <reason>`.
- `context.md`: free prose; `[text](/path.md)` links knowledge entries; backticked paths count as touched files for retrieval.
- `decisions.md`: entries `## <id>: <title>` followed by `date: YYYY-MM-DD`, optional `supersedes: <id>` and `outcome: true`, then the rationale and alternatives. Append-only. Blank lines between the heading and the metadata lines are accepted when reading; `prep fmt` removes them (canonical form: metadata directly under the heading).
- Records written only by commands: `baselines/<YYYYMMDD-HHMMSS>.md` (by, at, kind, ack; body = requirement; baseline names are per issue and stay timestamps), `ready.md` (by, at, baseline), `claim.md` (by, at), `resolution.md` (outcome, by, at, reason, evidence, documentation.entries / no_impact, dod, dod_opt_outs).
- Canonical form: LF, no trailing whitespace, single blank lines outside code fences, frontmatter key order fixed, one field per line, checkboxes as `- [x]`. `prep fmt` normalizes; CI runs `prep fmt --check`.
- Completion policy extensions within schema 1: project.md has optional `completion_policy` with verification, documentation, by and at; ready.md has optional `contract`; verification.md has an `evidence` frontmatter list and no body; resolution.md has optional `contract` and `verification`. Reports carry identities, history links, result, method, observation, provenance, actor/time, work/configuration scopes and a SHA-256 snapshot; automated reports also require the command run. Legacy omissions stay omissions. See [Completion policy](/features/completion-policy.md).
- Approval extensions: approvals.md has an approvals frontmatter list and no body, plus optional dod_revision and opt_out_revisions heads tracking CLI-mediated DoD/request changes; project.md optionally records completion_policy_history, ready.md previous_contracts, and resolution.md approvals. A bound opt-out carries approval_id. These additive fields are omitted on legacy records; detailed shapes and validity are in [Completion approvals](/features/completion-approvals.md).
- Knowledge-review extensions: `knowledge-reviews.md` has a `reviews` frontmatter list and no body. Each report binds a reported updated, unchanged or no-impact assessment to the ready contract, entry content revision where applicable, explicit work/config scopes, actor/time and fingerprint; IDs form a content-hash chain. `resolution.md` optionally snapshots the reports actually selected for completion. Legacy resolutions omit this field and remain readable. Applicability is derived on load; update/confirm bookkeeping does not change an entry revision. See [Completion policy](/features/completion-policy.md).

- Resolution review extensions: optional resolution generation; resolution-history.md stores earlier done records with notes and resolution-reviews.md stores a hash-chained human review list. Both require frontmatter and no body. depends_on_approval is a subset of ordinary local or cross-project dependencies. See [Resolution review](/features/resolution-review.md).
- `dependency-overrides.md` stores a hash-chained `overrides` frontmatter list: reference, source/local snapshot digests, explicit human actor/reason, timestamp and previous identity. Applicability is derived from fresh local source observations; removed/renamed references remain visible as retired stale overrides.
- Local storage metadata: `.prep/local/write.lock` serializes cooperating readers/writers; `.prep/.write-transaction.json` and sibling `.prep-tmp-*` stages support roll-forward recovery before records are loaded. These are recovery metadata, not domain records. See [Local write safety](/features/local-write-safety.md) for the bounded guarantee and conflicts.
