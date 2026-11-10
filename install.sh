#!/bin/sh

set -eu

REPOSITORY="scc-digitalhub/digitalhub-cli"
VERSION="${1:-latest}"
INSTALL_DIR="${DHCLI_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) echo "Unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) ARCH="amd64" ;;
  arm64 | aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  RELEASE_URL="https://github.com/$REPOSITORY/releases/latest/download"
else
  case "$VERSION" in
    v*) TAG="$VERSION" ;;
    *) TAG="v$VERSION" ;;
  esac
  RELEASE_URL="https://github.com/$REPOSITORY/releases/download/$TAG"
fi

ARCHIVE="dhcli-$OS-$ARCH.tar.gz"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

download() {
  URL="$1"
  DESTINATION="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "$DESTINATION"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$URL" -O "$DESTINATION"
  else
    echo "Install curl or wget to download dhcli" >&2
    exit 1
  fi
}

download "$RELEASE_URL/$ARCHIVE" "$TMP_DIR/$ARCHIVE"
download "$RELEASE_URL/checksums.txt" "$TMP_DIR/checksums.txt"

EXPECTED="$(awk -v file="$ARCHIVE" '$2 == file { print $1; exit }' "$TMP_DIR/checksums.txt")"
if [ -z "$EXPECTED" ]; then
  echo "No checksum found for $ARCHIVE" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "$TMP_DIR/$ARCHIVE" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "$TMP_DIR/$ARCHIVE" | awk '{ print $1 }')"
else
  echo "Install sha256sum or shasum to verify the downloaded archive" >&2
  exit 1
fi

if [ "$ACTUAL" != "$EXPECTED" ]; then
  echo "Checksum verification failed for $ARCHIVE" >&2
  exit 1
fi

mkdir -p "$TMP_DIR/extracted" "$INSTALL_DIR"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR/extracted" dhcli
install -m 755 "$TMP_DIR/extracted/dhcli" "$INSTALL_DIR/dhcli"

echo "Installed dhcli to $INSTALL_DIR/dhcli"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Add this directory to your PATH: $INSTALL_DIR" ;;
esac