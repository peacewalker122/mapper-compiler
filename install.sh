#!/bin/sh
# Install mapper-gen, mapper-gen-go, and mapper-gen-ts from mapper-compiler GitHub Releases.
#
# One-line install (Linux, macOS, Windows via Git Bash):
#   curl -fsSL https://raw.githubusercontent.com/peacewalker122/mapper-compiler/main/install.sh | bash
#
# Pin the version and/or install directory:
#   curl -fsSL https://raw.githubusercontent.com/peacewalker122/mapper-compiler/main/install.sh \
#     | MAPPER_VERSION=v0.1.0 MAPPER_INSTALL_DIR="$HOME/.local/bin" bash
#
# Or download first to inspect, then run:
#   curl -fsSL -o install.sh https://raw.githubusercontent.com/peacewalker122/mapper-compiler/main/install.sh
#   sh install.sh --version v0.1.0 --dir ~/.local/bin
set -eu

REPO="peacewalker122/mapper-compiler"
VERSION="${MAPPER_VERSION:-}"
INSTALL_DIR="${MAPPER_INSTALL_DIR:-}"

usage() {
  cat <<USAGE
Usage: install.sh [--version vX.Y.Z] [--dir DIR]

Installs mapper-gen, mapper-gen-go, and mapper-gen-ts from mapper-compiler GitHub Releases.

Options:
  --version vX.Y.Z   Release tag to install (default: latest release,
                     or \$MAPPER_VERSION when piped).
  --dir DIR          Install directory (default: /usr/local/bin when writable,
                     otherwise \$HOME/.local/bin; or \$MAPPER_INSTALL_DIR).
  -h, --help         Show this help.
USAGE
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:?--version requires a value}"; shift 2 ;;
    --dir) INSTALL_DIR="${2:?--dir requires a value}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "error: unknown argument: $1" >&2; usage >&2; exit 1 ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command not found: $1" >&2
    exit 1
  fi
}

download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$dest" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$dest" "$url"
  else
    echo "error: need curl or wget to download releases" >&2
    exit 1
  fi
}

# --- Detect platform (matches GoReleaser archive naming) ---
os="$(uname -s)"
case "$os" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
  *) echo "error: unsupported OS: $os" >&2; exit 1 ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "error: unsupported architecture: $arch" >&2; exit 1 ;;
esac

if [ "$OS" = "windows" ]; then EXT="zip"; else EXT="tar.gz"; fi
need_cmd uname
need_cmd mktemp
if [ "$EXT" = "zip" ]; then need_cmd unzip; else need_cmd tar; fi

# --- Resolve version (tags look like v0.1.0) ---
if [ -z "$VERSION" ]; then
  echo "Resolving latest mapper-compiler release..."
  api_json="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || true)"
  VERSION="$(printf '%s' "$api_json" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$VERSION" ]; then
    echo "error: could not resolve latest release; set MAPPER_VERSION (e.g. MAPPER_VERSION=v0.1.0) or pass --version" >&2
    exit 1
  fi
fi
echo "Installing mapper-compiler ${VERSION} for ${OS}/${ARCH}..."

FILE_VER="$(printf '%s' "$VERSION" | sed 's/^v//')"
ARCHIVE="mapper-compiler_${FILE_VER}_${OS}_${ARCH}.${EXT}"
CHECKSUMS="mapper-compiler_${FILE_VER}_checksums.txt"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

# --- Default install directory ---
if [ -z "$INSTALL_DIR" ]; then
  if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi

# --- Download archive + checksums ---
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT INT TERM
download "${BASE_URL}/${ARCHIVE}" "${TMPDIR}/${ARCHIVE}"
download "${BASE_URL}/${CHECKSUMS}" "${TMPDIR}/${CHECKSUMS}"

# --- Verify SHA-256 checksum ---
echo "Verifying checksum..."
(cd "$TMPDIR" && grep -F " ${ARCHIVE}" "${CHECKSUMS}" > expected.sha256)
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$TMPDIR" && sha256sum -c expected.sha256)
elif command -v shasum >/dev/null 2>&1; then
  (cd "$TMPDIR" && shasum -a 256 -c expected.sha256)
else
  echo "error: need sha256sum or shasum to verify the download" >&2
  exit 1
fi

# --- Extract (archives wrap binaries in a top-level directory) ---
mkdir -p "${TMPDIR}/extracted"
if [ "$EXT" = "zip" ]; then
  unzip -q -o "${TMPDIR}/${ARCHIVE}" -d "${TMPDIR}/extracted"
else
  tar -xzf "${TMPDIR}/${ARCHIVE}" -C "${TMPDIR}/extracted"
fi
MAPPER_GEN="$(find "${TMPDIR}/extracted" -name mapper-gen -type f | head -n 1)"
MAPPER_GEN_GO="$(find "${TMPDIR}/extracted" -name mapper-gen-go -type f | head -n 1)"
MAPPER_GEN_TS="$(find "${TMPDIR}/extracted" -name mapper-gen-ts -type f | head -n 1)"
if [ -z "$MAPPER_GEN" ] || [ -z "$MAPPER_GEN_GO" ] || [ -z "$MAPPER_GEN_TS" ]; then
  echo "error: archive did not contain mapper-gen, mapper-gen-go, and mapper-gen-ts" >&2
  exit 1
fi

# --- Install ---
mkdir -p "$INSTALL_DIR"
cp "$MAPPER_GEN" "$INSTALL_DIR/mapper-gen"
cp "$MAPPER_GEN_GO" "$INSTALL_DIR/mapper-gen-go"
cp "$MAPPER_GEN_TS" "$INSTALL_DIR/mapper-gen-ts"
chmod +x "$INSTALL_DIR/mapper-gen" "$INSTALL_DIR/mapper-gen-go" "$INSTALL_DIR/mapper-gen-ts"

# --- Smoke test (mapper-gen with no args prints usage and exits 1) ---
if "$INSTALL_DIR/mapper-gen" 2>&1 | grep -q "usage: mapper-gen"; then
  echo "Installed mapper-gen, mapper-gen-go, and mapper-gen-ts to ${INSTALL_DIR}"
else
  echo "error: installed mapper-gen failed to run" >&2
  exit 1
fi

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Note: ${INSTALL_DIR} is not on PATH. Add it, e.g.: export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
esac
