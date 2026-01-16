#!/usr/bin/env bash
set -euo pipefail

# --- Load configuration ------------------------------------------------------

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ ! -f "$SCRIPT_DIR/release.env" ]; then
  echo "ERROR: release.env not found" >&2
  exit 1
fi

source "$SCRIPT_DIR/release.env"

# --- Helpers -----------------------------------------------------------------

fail() {
  echo "ERROR: $1" >&2
  exit 1
}

info() {
  echo "==> $1"
}

# --- Preconditions -----------------------------------------------------------

cd "$PROJECT_ROOT"

CURRENT_TAG="$(git describe --tags --exact-match 2>/dev/null || true)"
[ -z "$CURRENT_TAG" ] && fail "no exact git tag found on HEAD"

info "current tag: $CURRENT_TAG"

# --- Fetch latest release from GitHub ---------------------------------------

info "fetching latest release from GitHub"

API_URL="https://api.github.com/repos/${OWNER_REPO}/releases/latest"
LATEST_TAG="$(curl -s "$API_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')"

[ -z "$LATEST_TAG" ] && fail "failed to fetch latest release tag from GitHub"

info "latest GitHub release: $LATEST_TAG"

# --- Compare versions --------------------------------------------------------

if [ "$CURRENT_TAG" != "$LATEST_TAG" ]; then
  fail "version mismatch: local=$CURRENT_TAG, GitHub=$LATEST_TAG"
fi

# --- Download and verify host-native binary ----------------------------------

info "downloading host-native binary from GitHub"

HOST_GOOS="$(go env GOOS)"
HOST_GOARCH="$(go env GOARCH)"
HOST_EXT=""

if [ "$HOST_GOOS" = "windows" ]; then
  HOST_EXT=".exe"
fi

BINARY_NAME="${BINARY}-${HOST_GOOS}-${HOST_GOARCH}${HOST_EXT}"
CHECKSUM_NAME="${BINARY_NAME}.sha256"

DOWNLOAD_DIR="$(mktemp -d)"
trap "rm -rf '$DOWNLOAD_DIR'" EXIT

DOWNLOAD_URL="https://github.com/${OWNER_REPO}/releases/download/${LATEST_TAG}/${BINARY_NAME}"
CHECKSUM_URL="https://github.com/${OWNER_REPO}/releases/download/${LATEST_TAG}/${CHECKSUM_NAME}"

curl -sL -o "${DOWNLOAD_DIR}/${BINARY_NAME}" "$DOWNLOAD_URL" ||
  fail "failed to download binary from $DOWNLOAD_URL"

curl -sL -o "${DOWNLOAD_DIR}/${CHECKSUM_NAME}" "$CHECKSUM_URL" ||
  fail "failed to download checksum from $CHECKSUM_URL"

# --- Verify checksum ---------------------------------------------------------

info "verifying checksum"

(cd "$DOWNLOAD_DIR" && sha256sum -c "$CHECKSUM_NAME") ||
  fail "checksum verification failed"

# --- Verify version ----------------------------------------------------------

info "verifying binary version"

chmod +x "${DOWNLOAD_DIR}/${BINARY_NAME}"
RAW_VERSION="$("${DOWNLOAD_DIR}/${BINARY_NAME}" --version)"
BINARY_VERSION="${RAW_VERSION##* }"

if [ "$BINARY_VERSION" != "$LATEST_TAG" ]; then
  fail "version mismatch in downloaded binary (expected $LATEST_TAG, got $RAW_VERSION)"
fi

# --- Success -----------------------------------------------------------------

cat <<EOF

✅ Post-release verification complete.

Verified:
  - Latest GitHub release tag matches local tag: $LATEST_TAG
  - Downloaded binary checksum is valid
  - Binary reports correct version: $BINARY_VERSION

EOF
