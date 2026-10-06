#!/bin/sh
# Install prep from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh | sh
#
# PREP_VERSION=v0.1.0 picks a version (default: latest), PREP_BINDIR the
# install directory (default: ~/.local/bin). The archive is verified
# against the release's checksums.txt before anything is installed.
set -eu

repo="fluid-movement/prep"
base="${PREP_RELEASE_URL:-https://github.com/$repo/releases}"
bindir="${PREP_BINDIR:-$HOME/.local/bin}"

fail() { echo "prep install: $*" >&2; exit 1; }

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) fail "unsupported system $(uname -s); download a release from https://github.com/$repo/releases" ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) fail "unsupported architecture $(uname -m)" ;;
esac

if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1" -o "$2"; }
  resolve() { curl -fsSLI -o /dev/null -w '%{url_effective}' "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -q "$1" -O "$2"; }
  resolve() { wget -q --max-redirect=5 -S --spider "$1" 2>&1 | sed -n 's/^ *[Ll]ocation: //p' | tail -1; }
else
  fail "needs curl or wget"
fi

version="${PREP_VERSION:-}"
if [ -z "$version" ]; then
  version=$(resolve "$base/latest" | sed 's|.*/tag/||' | tr -d '\r')
  [ -n "$version" ] || fail "could not find the latest release"
fi
case "$version" in v*) ;; *) version="v$version" ;; esac

archive="prep_${version#v}_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "prep install: downloading $version for $os/$arch"
fetch "$base/download/$version/$archive" "$tmp/$archive" || fail "no $archive in release $version"
fetch "$base/download/$version/checksums.txt" "$tmp/checksums.txt" || fail "release $version has no checksums.txt"

want=$(awk -v f="$archive" '$2 == f || $2 == "*" f { print $1 }' "$tmp/checksums.txt")
[ -n "$want" ] || fail "checksums.txt has no entry for $archive"
if command -v sha256sum >/dev/null 2>&1; then
  got=$(sha256sum "$tmp/$archive" | awk '{print $1}')
else
  got=$(shasum -a 256 "$tmp/$archive" | awk '{print $1}')
fi
[ "$got" = "$want" ] || fail "checksum mismatch for $archive; nothing was installed"

tar -xzf "$tmp/$archive" -C "$tmp" prep
mkdir -p "$bindir"
install -m 0755 "$tmp/prep" "$bindir/prep"
echo "prep install: installed $("$bindir/prep" version) to $bindir/prep"
case ":$PATH:" in
  *":$bindir:"*) ;;
  *)
    # Name the rc file of the user's shell so the hint can be pasted as is.
    case "$bindir" in "$HOME"/*) dir="\$HOME${bindir#"$HOME"}" ;; *) dir="$bindir" ;; esac
    case "$(basename "${SHELL:-sh}")" in
      zsh) line="echo 'export PATH=\"$dir:\$PATH\"' >> ~/.zshrc" ;;
      bash)
        # macOS terminals start login shells, which read ~/.bash_profile.
        rc="~/.bashrc"
        [ "$os" = darwin ] && rc="~/.bash_profile"
        line="echo 'export PATH=\"$dir:\$PATH\"' >> $rc" ;;
      fish) line="fish_add_path $bindir" ;;
      *) line="echo 'export PATH=\"$dir:\$PATH\"' >> ~/.profile" ;;
    esac
    echo "prep install: $bindir is not on your PATH yet. Add it with:"
    echo "  $line"
    echo "then open a new terminal."
    ;;
esac
