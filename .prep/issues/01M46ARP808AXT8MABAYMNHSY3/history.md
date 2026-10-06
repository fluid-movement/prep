- 2026-10-06T06:48:44Z edited by claude-code/2.1.289: tags

- 2026-10-06T08:39:55Z edited by claude-code/2.1.289: tags, requirement

- 2026-10-06T08:46:47Z claude-code/2.1.289: goreleaser check validated the configuration; a snapshot build produced prep_<version>_<os>_<arch> archives for all six platforms plus checksums.txt, matching update.ArchiveName and install.sh.

- 2026-10-06T08:46:47Z claude-code/2.1.289: install.sh tested against a local fake release (latest via redirect, PREP_VERSION, checksum mismatch refusal) and checked with dash; the PATH hint names the rc file of the user's shell after the user asked about ~/.local/bin.
