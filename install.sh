#!/bin/sh
set -e

# rlink POSIX installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/quaywin/rlink/main/install.sh | bash

REPO_OWNER="quaywin"
REPO_NAME="rlink"
BINARY_NAME="rlink"

# Terminal Output Utilities
info() {
    printf "\033[34m[INFO]\033[0m %s\n" "$1"
}

success() {
    printf "\033[32m[SUCCESS]\033[0m %s\n" "$1"
}

error() {
    printf "\033[31m[ERROR]\033[0m %s\n" "$1" >&2
    exit 1
}

# 1. Detect OS
OS_TYPE="$(uname -s)"
case "${OS_TYPE}" in
    Darwin*)  OS="darwin" ;;
    Linux*)   OS="linux" ;;
    *)        error "Unsupported Operating System: ${OS_TYPE}. Only macOS and Linux are supported." ;;
esac

# 2. Detect Architecture
ARCH_TYPE="$(uname -m)"
case "${ARCH_TYPE}" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)              error "Unsupported Architecture: ${ARCH_TYPE}. Only amd64 and arm64 are supported." ;;
esac

info "Detected platform: ${OS}/${ARCH}"

# 3. Fetch latest release version from GitHub API
API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
info "Fetching latest release from ${API_URL}..."

LATEST_TAG=$(curl -sSL -H "Accept: application/vnd.github+json" "${API_URL}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${LATEST_TAG}" ]; then
    error "Could not determine latest tag release for ${REPO_OWNER}/${REPO_NAME}."
fi

VERSION="${LATEST_TAG#v}"
info "Target release version: v${VERSION}"

# Construct Archive Name & Download URL
ARCHIVE_NAME="${REPO_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_TAG}/${ARCHIVE_NAME}"

# 4. Download to temporary directory
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'rlink')
trap 'rm -rf "${TMP_DIR}"' EXIT

info "Downloading ${DOWNLOAD_URL}..."
if [ -t 1 ] || [ -t 2 ]; then
    if ! curl -# -fSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${ARCHIVE_NAME}"; then
        error "Failed to download archive from ${DOWNLOAD_URL}"
    fi
else
    if ! curl -sSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${ARCHIVE_NAME}"; then
        error "Failed to download archive from ${DOWNLOAD_URL}"
    fi
fi

info "Extracting ${ARCHIVE_NAME}..."
tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "${TMP_DIR}"

if [ ! -f "${TMP_DIR}/${BINARY_NAME}" ]; then
    error "Binary '${BINARY_NAME}' not found inside archive."
fi

# 5. Installation Strategy
INSTALL_DIR=""
NEED_PATH_WARN=0

if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    INSTALL_DIR="$HOME/.local/bin"
elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="$HOME/bin"
    mkdir -p "${INSTALL_DIR}"
fi

info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
rm -f "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || true
cp "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*)
        NEED_PATH_WARN=0
        ;;
    *)
        NEED_PATH_WARN=1
        ;;
esac

success "rlink v${VERSION} has been successfully installed!"

if [ "${NEED_PATH_WARN}" -eq 1 ]; then
    printf "\n"
    printf "\033[33m[NOTE]\033[0m Directory '%s' is not in your current PATH.\n" "${INSTALL_DIR}"
    printf "       To use rlink globally, add the following line to your shell configuration profile (~/.zshrc or ~/.bashrc):\n\n"
    printf "       export PATH=\"%s:\$PATH\"\n\n" "${INSTALL_DIR}"
fi

printf "Run 'rlink setup' to get started.\n"
