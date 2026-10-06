## D1: Project-local configuration, no plugin yet
date: 2026-10-05

The hook and commands live in this repository's `.claude/` directory. A Claude Code plugin through a marketplace is the intended distribution for other projects, but it has to check binary compatibility, which depends on the install and update flow (01M46ARN8RW78DATMQV6P9NVYK). Alternatives: a plugin now (premature without a release channel); `prep init` writing `.claude/` files (couples the core to one harness).

## D2: Hook runs prime only when the binary is present
date: 2026-10-05

The hook is `command -v prep >/dev/null 2>&1 && prep prime || true`. Installing the binary in cloud sessions waits for the install flow; until then the skill's "Without the binary" section covers those sessions. Alternative: download a binary in the hook (no release artifacts exist yet).
