#!/bin/sh

set -eu

repository="${PB_AGENT_REPOSITORY:-anirudh-777/pb-agent}"
version="${PB_AGENT_VERSION:-${1:-}}"

say() {
  printf '%s\n' "$*"
}

fail() {
  printf 'pb-agent installer: %s\n' "$*" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

case "$(uname -s)" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *) fail "unsupported operating system; use a release archive on Windows" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) fail "unsupported architecture: $(uname -m)" ;;
esac

if [ -z "$version" ]; then
  version="$(
    curl -fsSL \
      -H "Accept: application/vnd.github+json" \
      "https://api.github.com/repos/$repository/releases?per_page=1" |
      awk -F '"' '/"tag_name":/ { print $4; exit }'
  )"
  [ -n "$version" ] || fail "could not determine the newest GitHub release"
fi

tag="$version"
case "$tag" in
  v*) version="${tag#v}" ;;
  *) tag="v$tag" ;;
esac

archive="pb-agent_${version}_${os}_${arch}.tar.gz"
release_url="https://github.com/$repository/releases/download/$tag"

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/pb-agent-install.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

say "Downloading pb-agent $version for $os/$arch..."
curl -fsSL "$release_url/$archive" -o "$tmp_dir/$archive"
curl -fsSL "$release_url/checksums.txt" -o "$tmp_dir/checksums.txt"

expected="$(
  awk -v archive="$archive" '$2 == archive { print $1; exit }' \
    "$tmp_dir/checksums.txt"
)"
[ -n "$expected" ] || fail "release checksum is missing for $archive"

if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp_dir/$archive" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$tmp_dir/$archive" | awk '{ print $1 }')"
else
  fail "sha256sum or shasum is required to verify the download"
fi

[ "$actual" = "$expected" ] || fail "checksum verification failed"

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"
[ -f "$tmp_dir/pb-agent" ] || fail "release archive does not contain pb-agent"

if [ -n "${PB_AGENT_INSTALL_DIR:-}" ]; then
  install_dir="$PB_AGENT_INSTALL_DIR"
elif [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
  install_dir="/usr/local/bin"
else
  install_dir="${HOME:?HOME is required}/.local/bin"
fi

mkdir -p "$install_dir"
[ -w "$install_dir" ] || fail "$install_dir is not writable; set PB_AGENT_INSTALL_DIR"

install -m 0755 "$tmp_dir/pb-agent" "$install_dir/pb-agent"

say "Installed pb-agent $version to $install_dir/pb-agent"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) say "Add $install_dir to PATH before running pb-agent." ;;
esac
