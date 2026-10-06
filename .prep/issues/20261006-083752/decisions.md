## D1: Generate ULIDs with the standard library, entropy injectable
date: 2026-10-06

A ULID is 48 bits of milliseconds and 80 bits from `crypto/rand`, encoded in Crockford base32. That is a few dozen lines, so prep implements it in `internal/domain` instead of adding a dependency (Stack: few dependencies). Within one tree, a new ULID must sort after every existing ID created in the same millisecond, so generation retries or increments like `nextStamp` does today. The random source is injectable so tests stay deterministic.

Alternatives considered: `github.com/oklog/ulid` (a dependency for code this small); a timestamp with a random suffix (readable, but the user chose ULIDs).

## D2: Keep timestamp IDs valid; order by creation time
date: 2026-10-06

Both formats are valid issue IDs. A helper returns an ID's creation time: the timestamp for the old format, the 48-bit time for a ULID. Every place that orders issues by ID (`IDs`, query tie-breaks, tree order) compares creation times, then the ID string. This keeps existing directories, references and commit messages unchanged (the user's choice of the recommended option).

Alternative considered: migrating every issue to a ULID once. One format, but every directory, every parent and dependency reference, and every ID in past commit messages and notes would change or dangle.
