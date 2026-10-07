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
confirmed_commit: b02849e589a03a17c2fd8730d5445263307734de
---

# Storage format

Applies when editing files in `.prep` by hand or changing the markdown adapter. Schema version 1.

- `project.md`: frontmatter `schema`; bullets under `## Definition of Done` are the project DoD.
- `config.yaml.dist` (committed, the project's default) and `config.yaml` (each user's own, ignored by `.prep/.gitignore`): the own file replaces the dist file as a whole when it exists. Both hold leading `#` comment lines, `views` (name → query flags, in tab order), `theme` (a theme name) and `themes` (custom themes: `base` and token → `"#RRGGBB"` pairs under the theme name; see [Palette and themes](/components/palette.md)). The TUI settings write the own file in this form, creating it from the dist file, keeping the leading comments; other comments are not kept. Unknown keys, such as the removed `commit_mode`, are errors (P003).
- `issues/<id>/`: the ID is a ULID (26 characters of Crockford base32, upper case: 48 bits of creation milliseconds and 80 random bits), so IDs created on different branches do not collide and sort by creation time as strings. Any other directory name is I001. Allowed entries are issue.md, acceptance.md, context.md, decisions.md, history.md, findings.md (research only), baselines/, ready.md, claim.md, resolution.md, attachments/. Anything else is an error.
- `issue.md`: frontmatter `title`, `kind`, `parent`, `depends_on`, `tags`, `priority` (in that order, nothing else); tags are lowercase letters, digits and `. _ - /`; priority is `critical`, `high` or `low`, absent meaning medium (medium is never written; added within schema 1). The body is the requirement; `## Open questions` holds unresolved questions. A `##` heading appears at most once per record file (`issue.md`, `acceptance.md`, `context.md`, `findings.md`; headings in code fences do not count): prep check reports a repeat as I028, an error for Open questions, which prep reads only the first copy of, and a warning otherwise.

- `acceptance.md`: `- [ ]` / `- [x]` criteria; optional `## Definition of Done` with additions and `- opt-out: <item> — <reason>`.
- `context.md`: free prose; `[text](/path.md)` links knowledge entries; backticked paths count as touched files for retrieval.
- `decisions.md`: entries `## <id>: <title>` followed by `date: YYYY-MM-DD`, optional `supersedes: <id>` and `outcome: true`, then the rationale and alternatives. Append-only. Blank lines between the heading and the metadata lines are accepted when reading; `prep fmt` removes them (canonical form: metadata directly under the heading).
- Records written only by commands: `baselines/<YYYYMMDD-HHMMSS>.md` (by, at, kind, ack; body = requirement; baseline names are per issue and stay timestamps), `ready.md` (by, at, baseline), `claim.md` (by, at), `resolution.md` (outcome, by, at, reason, evidence, documentation.entries / no_impact, dod, dod_opt_outs).
- Canonical form: LF, no trailing whitespace, single blank lines outside code fences, frontmatter key order fixed, one field per line, checkboxes as `- [x]`. `prep fmt` normalizes; CI runs `prep fmt --check`.
