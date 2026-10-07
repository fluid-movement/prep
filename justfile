# Common development tasks for prep. Run `just` to list them.

version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
ldflags := "-s -w -X github.com/fluid-movement/prep/internal/cli.Version=" + version
# Where `just install` places the binary; override with `just bindir=/usr/local/bin install`.
bindir := env("PREP_BINDIR", home_directory() / ".local" / "bin")

[private]
default:
    @just --list

# Build ./prep with the version stamped in
build:
    go build -ldflags '{{ldflags}}' -o prep ./cmd/prep

# Build and copy the binary into bindir (default ~/.local/bin), then point
# the harness integrations (the Claude Code plugin) at this checkout
install: build
    mkdir -p '{{bindir}}'
    install -m 0755 prep '{{bindir}}/prep'
    @echo "installed {{bindir}}/prep ({{version}})"
    @case ":$PATH:" in *":{{bindir}}:"*) ;; *) echo "note: {{bindir}} is not on your PATH";; esac
    @# A development build has no release to pin the plugin to, so prep setup
    @# installs it from this checkout; harnesses not set up are left alone.
    @PREP_PLUGIN_SOURCE='{{justfile_directory()}}' '{{bindir}}/prep' setup --refresh \
        && echo "harness integrations refreshed from this checkout; restart Claude Code or run /reload-plugins" \
        || echo "note: prep setup --refresh failed; the binary is installed"

# Remove the installed binary
uninstall:
    rm -f '{{bindir}}/prep'

# Run all tests, including the contract corpus
test:
    go test ./...

# Static checks: gofmt and go vet
lint:
    test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)
    go vet ./...

# Format Go code and .prep files
fmt:
    gofmt -w .
    go run ./cmd/prep fmt

# Validate this repository's own .prep tree
check:
    go run ./cmd/prep check
    go run ./cmd/prep fmt --check

# Everything CI runs
ci: lint test check

# Copy the embedded skill to the static copies: cloud sessions and the Claude Code plugin
sync-skill:
    cp internal/cli/skill.md .claude/skills/prep/SKILL.md
    cp internal/cli/skill.md plugins/claude-code/skills/prep/SKILL.md

# Prepare a release: set the plugin version, commit and tag; push the tag yourself
release version:
    #!/usr/bin/env sh
    set -eu
    case "{{version}}" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo "version must look like v1.2.3" >&2; exit 1 ;; esac
    if [ -n "$(git status --porcelain)" ]; then echo "commit or stash your changes first" >&2; exit 1; fi
    v="{{version}}"; v="${v#v}"
    for f in plugins/claude-code/.claude-plugin/plugin.json .claude-plugin/marketplace.json; do
        sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"$v\"/" "$f" && rm "$f.bak"
    done
    go test ./...
    git commit -qam "Release {{version}}"
    git tag -a "{{version}}" -m "prep {{version}}"
    echo "tagged {{version}}; publish it with: git push origin main {{version}}"

# Remove build output
clean:
    rm -f prep
