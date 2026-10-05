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

# Build and copy the binary into bindir (default ~/.local/bin)
install: build
    mkdir -p '{{bindir}}'
    install -m 0755 prep '{{bindir}}/prep'
    @echo "installed {{bindir}}/prep ({{version}})"
    @case ":$PATH:" in *":{{bindir}}:"*) ;; *) echo "note: {{bindir}} is not on your PATH";; esac

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

# Copy the embedded skill to the static copy agents read
sync-skill:
    cp internal/cli/skill.md .claude/skills/prep/SKILL.md

# Remove build output
clean:
    rm -f prep
