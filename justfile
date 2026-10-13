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
# the harness integrations at this checkout
install: build
    mkdir -p '{{bindir}}'
    install -m 0755 prep '{{bindir}}/prep'
    @echo "installed {{bindir}}/prep ({{version}})"
    @case ":$PATH:" in *":{{bindir}}:"*) ;; *) echo "note: {{bindir}} is not on your PATH";; esac
    @# Development builds have no release tag to pin plugins to, so prep setup
    @# installs them from this checkout; harnesses not set up are left alone.
    @PREP_PLUGIN_SOURCE='{{justfile_directory()}}' PREP_CODEX_PLUGIN_SOURCE='{{justfile_directory()}}' '{{bindir}}/prep' setup --refresh \
        && echo "harness integrations refreshed from this checkout; restart the harness and review any new Codex hooks in /hooks" \
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

# One-time login for the agent behavior tests: runs `claude setup-token` (opens a browser), then keeps the token it prints in ~/.config/prep/eval-token (mode 600)
eval-login:
    #!/usr/bin/env sh
    set -eu
    claude setup-token
    printf 'Paste the token printed above (input is hidden): '
    stty -echo; read -r token; stty echo; printf '\n'
    [ -n "$token" ] || { echo "no token given" >&2; exit 1; }
    dir="${XDG_CONFIG_HOME:-$HOME/.config}/prep"
    mkdir -p "$dir"
    ( umask 077; printf '%s' "$token" > "$dir/eval-token" )
    echo "saved $dir/eval-token; try: just eval-smoke"

# Agent behavior tests, the cheap smoke subset (needs `just eval-login` once, or CLAUDE_CODE_OAUTH_TOKEN / ANTHROPIC_API_KEY)
eval-smoke *args: build
    ./prep eval run --smoke --models haiku --n 3 {{args}}

# Agent behavior tests, every case on the default model; add --models haiku,sonnet,opus for a model comparison
eval-full *args: build
    ./prep eval run --models sonnet --n 5 {{args}}

# Copy the embedded skill to static copies: cloud sessions and agent plugins
sync-skill:
    cp internal/cli/skill.md .claude/skills/prep/SKILL.md
    cp internal/cli/skill.md plugins/claude-code/skills/prep/SKILL.md
    cp internal/cli/skill.md plugins/codex/skills/prep/SKILL.md
    printf '\n## Codex integration\n\nThe Codex hooks export `PREP_SESSION` and `PREP_ACTOR=codex` for shell commands, so prep CLI activity is attributed to this agent automatically. If hooks are disabled or untrusted, run `prep prime` and follow this skill manually.\n' >> plugins/codex/skills/prep/SKILL.md

# Prepare a release: set the plugin version, commit and tag; push the tag yourself
release version:
    #!/usr/bin/env sh
    set -eu
    case "{{version}}" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo "version must look like v1.2.3" >&2; exit 1 ;; esac
    if [ -n "$(git status --porcelain)" ]; then echo "commit or stash your changes first" >&2; exit 1; fi
    v="{{version}}"; v="${v#v}"
    for f in plugins/claude-code/.claude-plugin/plugin.json .claude-plugin/marketplace.json plugins/codex/.codex-plugin/plugin.json .agents/plugins/marketplace.json; do
        sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"$v\"/" "$f" && rm "$f.bak"
    done
    # The version is all that changed in the plugin entry's scope; the entry
    # names no number, so confirm it rather than leave a drift warning.
    go run ./cmd/prep knowledge confirm /components/claude-code.md >/dev/null
    go run ./cmd/prep knowledge confirm /components/codex.md >/dev/null
    go test ./...
    git commit -qam "Release {{version}}"
    git tag -a "{{version}}" -m "prep {{version}}"
    echo "tagged {{version}}; publish it with: git push origin main {{version}}"

# Build the macOS desktop app into desktop/build/Prep.app (ad-hoc signed)
desktop:
    #!/usr/bin/env sh
    set -eu
    cd desktop
    swift build -c release --product PrepDesktop
    app=build/Prep.app
    rm -rf "$app"
    mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
    cp "$(swift build -c release --show-bin-path)/PrepDesktop" "$app/Contents/MacOS/"
    v='{{version}}'; v="${v#v}"; case "$v" in [0-9]*) ;; *) v=0.0.0 ;; esac
    sed "s/VERSION/${v%%-*}/g" Resources/Info.plist > "$app/Contents/Info.plist"
    rm -rf build/AppIcon.iconset
    "$app/Contents/MacOS/PrepDesktop" --render-icon build/AppIcon.iconset >/dev/null
    iconutil -c icns build/AppIcon.iconset -o "$app/Contents/Resources/AppIcon.icns"
    codesign --force --sign - "$app"
    echo "built desktop/$app; open it with: open desktop/$app --args <project dir>"

# Test the desktop app's PrepKit (against ./prep, built first)
desktop-test: build
    cd desktop && PREP_BIN='{{justfile_directory()}}/prep' swift test

# Render the desktop app offscreen to a PNG: just desktop-snapshot out.png --select <id>
desktop-snapshot out *args: build
    cd desktop && swift build --product PrepDesktop
    PREP_BIN='{{justfile_directory()}}/prep' desktop/.build/debug/PrepDesktop '{{justfile_directory()}}' --snapshot '{{absolute_path(out)}}' {{args}}

# Remove build output
clean:
    rm -f prep
    rm -rf desktop/build desktop/.build
