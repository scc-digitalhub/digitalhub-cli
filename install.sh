#!/bin/sh

set -eu

printf '%s\n' \
  '   ___  __ _________   ____' \
  '  / _ \/ // / ___/ /  /  _/' \
  ' / // / _  / /__/ /___/ /  ' \
  '/____/_//_/\___/____/___/  ' \
  '                           ' \
  '' >&2

REPOSITORY="scc-digitalhub/digitalhub-cli"
VERSION="latest"
INSTALL_DIR="${DHCLI_INSTALL_DIR:-$HOME/.local/bin}"
UPDATE="false"
VERSION_SET="false"

for ARG in "$@"; do
  case "$ARG" in
    --update) UPDATE="true" ;;
    --*) echo "Unknown option: $ARG" >&2; exit 2 ;;
    *)
      if [ "$VERSION_SET" = "true" ]; then
        echo "Only one version may be specified" >&2
        exit 2
      fi
      VERSION="$ARG"
      VERSION_SET="true"
      ;;
  esac
done

if [ -f "$INSTALL_DIR/dhcli" ] && [ "$UPDATE" != "true" ]; then
  echo "dhcli already exists at $INSTALL_DIR/dhcli. Use --update to replace it." >&2
  exit 1
fi

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

fetch() {
  URL="$1"
  PROGRESS="${2:-}"
  if command -v curl >/dev/null 2>&1; then
    if [ "$PROGRESS" = "progress" ]; then
      curl --progress-bar -fL "$URL"
    else
      curl -fsSL "$URL"
    fi
  elif command -v wget >/dev/null 2>&1; then
    if [ "$PROGRESS" = "progress" ]; then
      wget -O - "$URL"
    else
      wget -qO- "$URL"
    fi
  else
    echo "Install curl or wget to download dhcli" >&2
    exit 1
  fi
}

if [ "$VERSION" = "latest" ]; then
  LATEST_RELEASE_API="https://api.github.com/repos/$REPOSITORY/releases/latest"
  TAG="$(fetch "$LATEST_RELEASE_API" | sed -n 's/^[[:space:]]*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | sed -n '1p')"
  if [ -z "$TAG" ]; then
    echo "Could not determine the latest dhcli release version" >&2
    exit 1
  fi
else
  TAG="$VERSION"
fi

RELEASE_URL="https://github.com/$REPOSITORY/releases/download/$TAG"

echo "Installing dhcli version $TAG to $INSTALL_DIR..."


ARCHIVE="dhcli-$OS-$ARCH.tar.gz"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

download() {
  fetch "$1" progress > "$2"
}

echo "Downloading $ARCHIVE and checksums..."

download "$RELEASE_URL/$ARCHIVE" "$TMP_DIR/$ARCHIVE"
download "$RELEASE_URL/dhcli_${TAG}_checksums.txt" "$TMP_DIR/checksums.txt"

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

echo "Extracting $ARCHIVE..."

mkdir -p "$TMP_DIR/extracted" "$INSTALL_DIR"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR/extracted" dhcli


echo "Installing dhcli to $INSTALL_DIR..."
install -m 755 "$TMP_DIR/extracted/dhcli" "$INSTALL_DIR/dhcli"

echo "Installed dhcli $TAG to $INSTALL_DIR/dhcli"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) echo "Run 'dhcli --help' to get started." ;;
  *) echo "Add $INSTALL_DIR to your PATH to run dhcli." ;;
esac