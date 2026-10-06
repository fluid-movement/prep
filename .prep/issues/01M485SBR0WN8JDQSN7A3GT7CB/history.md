- 2026-10-06T11:24:35Z edited by claude-code/2.1.291: parent

- 2026-10-06T13:58:01Z edited by claude-code/2.1.291: title, requirement

- 2026-10-06T14:01:11Z edited by claude-code/2.1.291: requirement

- 2026-10-06T14:03:30Z edited by claude-code/2.1.291: requirement

- 2026-10-06T14:04:33Z edited by claude-code/2.1.291: requirement

- 2026-10-06T14:18:13Z claude-code/cloud: TUI goldens (internal/tui/testdata, internal/tui/ui/testdata) regenerated with -update: only the shown IDs changed, from time parts to the last 6 ULID characters; the full ULID in the detail header now wraps onto its own line at 110 columns. Fixtures seed the ID entropy with rand.ChaCha8.

- 2026-10-06T14:18:13Z claude-code/cloud: Contract corpus converted with the same throwaway script as the repository (deterministic random bits from a hash of the old ID). In valid/lifecycle a baseline name equals an issue ID (20261005-152228); the script leaves baseline: lines and baseline file names alone.

- 2026-10-06T14:18:13Z claude-code/cloud: Built against the old binary until the repository is converted: the new binary rejects timestamp directories (I001), so the code commit alone fails prep check in CI; the conversion commit right after it fixes that.

- 2026-10-06T14:19:06Z claude-code/cloud: Repository converted in e3437da; the code is in 02533e3. Knowledge entries confirmed after re-checking architecture, lifecycle, tui-design-system and claude-code (their scoped files changed only in IDs and the skill text, nothing they describe).
