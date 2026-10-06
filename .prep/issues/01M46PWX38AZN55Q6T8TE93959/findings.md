# Cloud sessions without the prep binary

Researched from inside a Claude Code on the web session on 2026-10-06 (Claude Code 2.1.42, environment type `cloud_default`, this repository checked out at `c83ac06`).

## What the session offers at start (evidence)

- **System**: Ubuntu 24.04.5 LTS, Linux 6.18 x86_64 (`linux/amd64`), 4 CPUs, 15 GB RAM, running as root; `HOME=/root`. `CLAUDE_CODE_REMOTE=true` marks the session as remote.
- **Checkout**: a shallow clone (50 commits) of the repository on the session's branch; `.prep` is present, `prep` is not.
- **PATH**: `/root/.local/bin` is on the PATH (first entry) and writable; so are `/usr/local/bin` and the other PATH directories, since the session runs as root. A binary placed in `~/.local/bin` is usable without changing the PATH.
- **Tools**: `git`, `curl`, `wget`, `gh`; Go 1.24.7 at `/usr/local/go`, which downloads newer toolchains on demand (`GOTOOLCHAIN=auto`; `go build` here fetched go1.26.0, `go install` fetched go1.26.8). Node, Bun, Cargo, Maven and Gradle are installed too.
- **Network**: outbound HTTPS goes through an egress proxy that re-terminates TLS with a CA bundle the system trust store and `curl` already use. Measured:
  - `https://github.com/<owner>/<repo>/releases/download/...` (a public release asset, 1.5 MB): 200; release downloads work, so `install.sh` can fetch an archive and `checksums.txt`.
  - `https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh`: 200.
  - `https://api.github.com/repos/fluid-movement/prep`: 200, but the API for a repository outside the session's scope answers 403 ("GitHub access to this repository is not enabled for this session"). An installer must not depend on the GitHub API; `install.sh` resolves the latest release through the `releases/latest` redirect and is unaffected.
  - `proxy.golang.org` is on the proxy's allow list: `go install github.com/fluid-movement/prep/cmd/prep@latest` installed a working binary in 34 s (including the toolchain download).
  - The network policy is chosen per environment (none, trusted package registries, custom allowed domains, full); a stricter policy than this session's can block GitHub downloads, which is why a fallback stays necessary.
- **Hooks and setup**:
  - User settings come from the platform (`~/.claude/launcher-settings.json`, which registers its own Stop hook): user-scope hooks run, but only the platform's.
  - **No user plugins are installed** (`~/.claude/plugins` holds only an empty synced folder), so the prep plugin and its SessionStart hook do not run. The prep skill reached this session only because the repository carries `.claude/skills/prep/SKILL.md`: project skills load.
  - Project settings (`.claude/settings.json` in the repository, absent here) are the documented place for cloud SessionStart hooks: Claude Code's built-in session-start-hook skill describes them, with `$CLAUDE_PROJECT_DIR`, `$CLAUDE_ENV_FILE` (to export variables into the session) and `$CLAUDE_CODE_REMOTE`, and an async mode. Not observed running here, because this repository has none.
  - Environments also have a setup script, configured by the user in the environment settings on claude.ai, not in the repository; it runs before the session and is per environment, so a project cannot ship it.

## Recommendation

**Get the binary through a project SessionStart hook.** The repository carries the hook, so every cloud session of that project runs it without user setup. The hook, when `CLAUDE_CODE_REMOTE=true` and `prep` is missing:

1. installs the release pinned to the version that wrote the hook through `install.sh` (`PREP_VERSION=vX.Y.Z`, into `~/.local/bin`, checksum-verified); release downloads work under the default network policy and need no Go;
2. else, when Go is on the PATH, runs `go install github.com/fluid-movement/prep/cmd/prep@vX.Y.Z` (proxy.golang.org is reachable even under the package-registry policy);
3. then runs `prep prime --hook`, so the session starts with the same briefing as a local one; when both installs failed it prints one line saying the binary is missing and the session is in the fallback.

It must never fail or block session start (exit 0, short timeouts). Pinning keeps the binary in step with the schema the repository uses, the same reason the plugin marketplace is pinned. Trade-offs: the hook adds about a second (release) to half a minute (go install) to session start, async mode hides that but races with the first prep call; the pin must move when the project updates prep; a hook in the repository is code every contributor's sessions run, so it stays small and readable and only acts in remote sessions. Alternatives considered: the environment setup script (works but is per user and per environment, invisible to the project); the plugin hook (does not run in cloud sessions); building from source (only works in this repository); committing a binary (platform-specific, bloats the repository).

**The skill tries the install before falling back.** An agent that finds no binary runs the same two install commands itself before hand-editing; only when both fail does it edit files, and it tells the user so.

**What the fallback must guarantee.** Hand edits cannot be prevented, so they must be caught before they land: a CI check that installs the pinned release and runs `prep check` and `prep fmt --check` on every push and pull request. That guarantees structural validity and canonical form, and it must also reject records that only commands write when they are inconsistent (a `ready.md` against a baseline whose requirement differs from `issue.md`, a completion without the records its kind needs). This repository already runs `prep check` and `prep fmt --check` in CI (built from source); adopting projects need a ready-made version. Trade-off: the check catches errors after the push rather than at the edit, and gates the command skipped (for example claiming before ready) can only be caught when the resulting records show them.

## Follow-up issues

- 01M48RV2FR0XQAGCKEXRJFH7AE Install prep in cloud sessions through a project SessionStart hook
- 01M48RV3F0ZZB2MXBFR2R62P9J Skill: install the binary before falling back to hand edits
- 01M48RV4E8J5T57FN2P09SHG2W CI check that rejects invalid hand edits of .prep

Other harnesses (Codex cloud, Cursor background agents) follow once the Claude Code path is in place; the same three pieces (project-level install hook, install-first skill text, CI check) carry over wherever the harness runs repository-defined setup.
