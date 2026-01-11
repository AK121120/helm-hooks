#!/bin/bash
# Build script for helm-hooks
# Usage: ./scripts/build.sh [--all] [--vulnscan] [--version VERSION]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# Defaults
BUILD_ALL=false
VULN_SCAN=false
VERSION="${VERSION:-dev}"
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${YELLOW}→${NC} $1"; }
log_success() { echo -e "${GREEN}✓${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --all) BUILD_ALL=true; shift ;;
        --vulnscan) VULN_SCAN=true; shift ;;
        --version) VERSION="$2"; shift 2 ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --all        Build for all platforms"
            echo "  --vulnscan   Run vulnerability scan"
            echo "  --version V  Set version (default: dev)"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Go binary
GO=$(command -v go 2>/dev/null || echo /usr/local/go/bin/go)

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.GitCommit=$GIT_COMMIT"
LDFLAGS="$LDFLAGS -X main.BuildDate=$BUILD_DATE"

log_info "Building helm-hooks v${VERSION}"
echo "  Git Commit: $GIT_COMMIT"
echo "  Build Date: $BUILD_DATE"
echo

# Run tests first
log_info "Running tests..."
$GO test ./... -v
log_success "Tests passed"
echo

# Vulnerability scan
if [[ "$VULN_SCAN" == "true" ]]; then
    log_info "Running vulnerability scan..."
    if command -v govulncheck &>/dev/null; then
        govulncheck ./...
        log_success "No vulnerabilities found"
    else
        log_info "Installing govulncheck..."
        $GO install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...
        log_success "No vulnerabilities found"
    fi
    echo
fi

# Build
if [[ "$BUILD_ALL" == "true" ]]; then
    log_info "Building for all platforms..."
    
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )
    
    for platform in "${platforms[@]}"; do
        GOOS="${platform%/*}"
        GOARCH="${platform#*/}"
        output="helm-hooks-${GOOS}-${GOARCH}"
        [[ "$GOOS" == "windows" ]] && output+=".exe"
        
        log_info "Building $output..."
        GOOS=$GOOS GOARCH=$GOARCH $GO build -ldflags "$LDFLAGS" -o "$output" ./cmd/helm-hooks
    done
    
    log_success "Built all platforms"
else
    log_info "Building for current platform..."
    $GO build -ldflags "$LDFLAGS" -o helm-hooks ./cmd/helm-hooks
    log_success "Built ./helm-hooks"
fi

echo
log_success "Build complete!"
./helm-hooks version
