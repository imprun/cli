#!/bin/sh

set -eu

repository="imprun/cli"
release_base_url="${IMPRUN_RELEASE_BASE_URL:-https://github.com/${repository}/releases/download}"
version="${IMPRUN_VERSION:-}"
install_dir="${IMPRUN_INSTALL_DIR:-${XDG_BIN_HOME:-${HOME}/.local/bin}}"
require_signature="${IMPRUN_REQUIRE_SIGNATURE:-0}"

usage() {
  cat <<'EOF'
Install the Imprun CLI from a signed GitHub release.

Usage: install.sh [options]

Options:
  --version <version>       Install a specific version (for example, 0.3.1).
  --install-dir <path>      Install into this directory (default: ~/.local/bin).
  --require-signature       Fail unless cosign verifies the release checksum bundle.
  -h, --help                Show this help.

Environment variables:
  IMPRUN_VERSION
  IMPRUN_INSTALL_DIR
  IMPRUN_REQUIRE_SIGNATURE=1
EOF
}

fail() {
  printf 'imprun installer: %s\n' "$*" >&2
  exit 1
}

warn() {
  printf 'imprun installer: warning: %s\n' "$*" >&2
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || fail "--version requires a value"
      version=$2
      shift 2
      ;;
    --install-dir)
      [ "$#" -ge 2 ] || fail "--install-dir requires a value"
      install_dir=$2
      shift 2
      ;;
    --require-signature)
      require_signature=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown option: $1"
      ;;
  esac
done

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

if [ -z "$version" ]; then
  latest_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/${repository}/releases/latest") ||
    fail "could not resolve the latest stable release"
  version=${latest_url##*/}
fi

version=${version#v}
printf '%s\n' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z][0-9A-Za-z.-]*)?$' ||
  fail "invalid version: $version"

case "$(uname -s)" in
  Linux) platform=linux ;;
  Darwin) platform=darwin ;;
  *) fail "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) architecture=amd64 ;;
  arm64|aarch64) architecture=arm64 ;;
  *) fail "unsupported architecture: $(uname -m)" ;;
esac

tag="v${version}"
asset="imprun_${version}_${platform}_${architecture}.tar.gz"
release_url="${release_base_url%/}/${tag}"

temporary_dir=$(mktemp -d 2>/dev/null || mktemp -d -t imprun-install)
staged_target=""
cleanup() {
  if [ -n "$staged_target" ]; then
    rm -f "$staged_target"
  fi
  rm -rf "$temporary_dir"
}
trap cleanup EXIT HUP INT TERM

download() {
  source_url=$1
  destination=$2
  curl -fsSL "$source_url" -o "$destination" || fail "download failed: $source_url"
}

asset_path="${temporary_dir}/${asset}"
checksums_path="${temporary_dir}/checksums.txt"
bundle_path="${temporary_dir}/checksums.txt.sigstore.json"

download "${release_url}/${asset}" "$asset_path"
download "${release_url}/checksums.txt" "$checksums_path"

if command -v cosign >/dev/null 2>&1; then
  download "${release_url}/checksums.txt.sigstore.json" "$bundle_path"
  cosign verify-blob \
    --bundle "$bundle_path" \
    --certificate-identity "https://github.com/${repository}/.github/workflows/release.yml@refs/tags/${tag}" \
    --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
    "$checksums_path" >/dev/null || fail "release signature verification failed"
elif [ "$require_signature" = "1" ]; then
  fail "cosign is required when signature verification is mandatory"
else
  warn "cosign was not found; SHA-256 was verified but signer verification was skipped"
fi

checksum_matches=$(awk -v name="$asset" '$2 == name { count += 1 } END { print count + 0 }' "$checksums_path")
[ "$checksum_matches" -eq 1 ] || fail "the release checksum must contain exactly one entry for $asset"
expected_checksum=$(awk -v name="$asset" '$2 == name { print tolower($1) }' "$checksums_path")

if command -v sha256sum >/dev/null 2>&1; then
  actual_checksum=$(sha256sum "$asset_path" | awk '{ print tolower($1) }')
elif command -v shasum >/dev/null 2>&1; then
  actual_checksum=$(shasum -a 256 "$asset_path" | awk '{ print tolower($1) }')
else
  fail "sha256sum or shasum is required"
fi

[ "$actual_checksum" = "$expected_checksum" ] || fail "SHA-256 mismatch for $asset"

extract_dir="${temporary_dir}/extract"
mkdir -p "$extract_dir"
tar -xzf "$asset_path" -C "$extract_dir"
candidate="${extract_dir}/imprun"
[ -f "$candidate" ] || fail "release archive does not contain imprun"
chmod 0755 "$candidate"

candidate_version=$("$candidate" --version 2>&1) || fail "downloaded executable did not run"
case "$candidate_version" in
  "$version"|"imprun $version") ;;
  *) fail "downloaded executable reported an unexpected version: $candidate_version" ;;
esac

mkdir -p "$install_dir"
staged_target="${install_dir}/.imprun.new.$$"
cp "$candidate" "$staged_target"
chmod 0755 "$staged_target"
"$staged_target" --version >/dev/null 2>&1 || fail "staged executable did not run"
mv -f "$staged_target" "${install_dir}/imprun"
staged_target=""

printf 'Installed imprun %s to %s\n' "$version" "${install_dir}/imprun"
case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *) printf 'Add %s to PATH, then run: imprun --version\n' "$install_dir" ;;
esac
