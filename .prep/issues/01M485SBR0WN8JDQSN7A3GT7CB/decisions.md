## D1: Generate ULIDs with the standard library, entropy injectable
date: 2026-10-06

A ULID is 48 bits of milliseconds and 80 bits from `crypto/rand`, encoded in Crockford base32. That is a few dozen lines, so prep implements it in `internal/domain` instead of adding a dependency (Stack: few dependencies). Within one tree, a new ULID must sort after every existing ID created in the same millisecond, so generation retries or increments like `nextStamp` does today. The random source is injectable so tests stay deterministic.

Alternatives considered: `github.com/oklog/ulid` (a dependency for code this small); a timestamp with a random suffix (readable, but the user chose ULIDs).

## D2: Keep timestamp IDs valid; order by creation time
date: 2026-10-06

Both formats are valid issue IDs. A helper returns an ID's creation time: the timestamp for the old format, the 48-bit time for a ULID. Every place that orders issues by ID (`IDs`, query tie-breaks, tree order) compares creation times, then the ID string. This keeps existing directories, references and commit messages unchanged (the user's choice of the recommended option).

Alternative considered: migrating every issue to a ULID once. One format, but every directory, every parent and dependency reference, and every ID in past commit messages and notes would change or dangle.

## D3: ULIDs only, from schema 2; prep migrate converts existing projects
date: 2026-10-06
supersedes: D2

The user ruled out mixed formats: prep is in development and this repository is its only user. Schema 2 accepts only ULIDs as issue IDs. The migration from schema 1 builds each issue's ULID from its timestamp (the 48-bit time part, random bits from the injectable source), so creation order is kept. It renames the issue directories and rewrites every occurrence of an old ID in `.prep` text files: frontmatter references, records, bodies, logs and knowledge entries. A project on schema 1 gets P002 until it is migrated. Baseline file names stay timestamps: they are per issue, not IDs.

Alternatives considered: keeping both formats valid (the previous D2), which the user ruled out; rewriting only frontmatter references, which would leave dangling IDs in requirements, contexts and logs.

## D4: ULIDs only, no migration: convert this repository once
date: 2026-10-06
supersedes: D3

The user's call: prep is unreleased and this repository is its only user, so prep ships no migration. The schema version stays 1; `ValidID` accepts only ULIDs, and timestamp directories become I001 errors. This repository is converted once by a throwaway script, run from the scratch area and not committed. For each issue it builds a ULID from the issue's timestamp (the time part) and random bits, renames the directory, and rewrites every occurrence of the old ID in `.prep` text files. Afterwards `prep check` is clean. Baseline file names stay timestamps: they are per issue, not IDs.

Alternatives considered: a schema 2 step in `prep migrate` (the previous D3), which is code kept forever for a single conversion; keeping both formats (D2), which the user ruled out.
