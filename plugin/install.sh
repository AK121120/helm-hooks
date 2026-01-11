#!/bin/bash
# Install script for helm-hooks (Helm 3 users)
# Downloads pre-built binary and installs to current directory or PATH

set -e

VERSION="${HELM_HOOKS_VERSION:-latest}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
esac

echo "Installing helm-hooks ${VERSION} for ${OS}/${ARCH}..."

# GitHub release URL (update this to your actual repo)
GITHUB_REPO="agk/helm-hooks"
BINARY_NAME="helm-hooks"

if [[ "$VERSION" == "latest" ]]; then
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"
fi

# Download binary
INSTALL_DIR="${HELM_HOOKS_INSTALL_DIR:-.}"

if command -v curl &> /dev/null; then
    curl -sSL -o "${INSTALL_DIR}/${BINARY_NAME}" "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -q -O "${INSTALL_DIR}/${BINARY_NAME}" "$DOWNLOAD_URL"
else
    echo "Error: curl or wget is required"
    exit 1
fi

chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo "Installed: ${INSTALL_DIR}/${BINARY_NAME}"
echo
echo "Usage:"
echo "  helm install myapp ./chart --post-renderer ${INSTALL_DIR}/${BINARY_NAME}"
echo
echo "Or add to PATH:"
echo "  sudo mv ${INSTALL_DIR}/${BINARY_NAME} /usr/local/bin/"
