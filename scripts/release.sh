#!/bin/bash
# Release script for helm-hooks
# Usage: ./scripts/release.sh -v VERSION [--dry-run] [--skip-push]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# Configuration
GITHUB_REPO="AK121120/helm-hooks"
PLUGIN_REPO="AK121120/helm-hooks-plugin"
PLUGIN_DIR="helm-hooks-plugin"

# Defaults
VERSION=""
DRY_RUN=false
SKIP_PUSH=false

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${YELLOW}→${NC} $1"; }
log_success() { echo -e "${GREEN}✓${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; exit 1; }
log_header() { echo -e "\n${BLUE}═══════════════════════════════════════${NC}"; echo -e "${BLUE}  $1${NC}"; echo -e "${BLUE}═══════════════════════════════════════${NC}\n"; }

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version) VERSION="$2"; shift 2 ;;
        --dry-run) DRY_RUN=true; shift ;;
        --skip-push) SKIP_PUSH=true; shift ;;
        -h|--help)
            cat << EOF
Usage: $0 -v VERSION [OPTIONS]

Options:
  -v, --version VERSION   Version to release (required, e.g., 0.1.0)
  --dry-run               Show what would be done without executing
  --skip-push             Build only, don't push to GitHub

Examples:
  $0 -v 0.1.0                     # Release v0.1.0
  $0 -v 0.1.0 --dry-run           # Preview release
  $0 -v 0.1.0 --skip-push         # Build only, no push
EOF
            exit 0
            ;;
        *) log_error "Unknown option: $1" ;;
    esac
done

# Validate version
if [[ -z "$VERSION" ]]; then
    log_error "Version is required. Use -v VERSION"
fi

if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Version must be in format MAJOR.MINOR.PATCH (e.g., 0.1.0)"
fi

TAG="v$VERSION"
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Check if tag already exists
if git tag -l | grep -q "^${TAG}$"; then
    log_error "Tag $TAG already exists"
fi

# Check plugin submodule exists
if [[ ! -d "$PLUGIN_DIR" ]]; then
    log_error "Plugin submodule not found: $PLUGIN_DIR"
fi

log_header "helm-hooks Release $TAG"
echo "Version:    $VERSION"
echo "Git Commit: $GIT_COMMIT"
echo "Build Date: $BUILD_DATE"
echo "Dry Run:    $DRY_RUN"
echo

#############################################
# Step 1: Run Tests
#############################################
log_header "Step 1: Running Tests"

if [[ "$DRY_RUN" == "true" ]]; then
    log_info "[DRY RUN] Would run tests"
else
    GO=$(command -v go 2>/dev/null || echo /usr/local/go/bin/go)
    $GO test ./... -v
    log_success "Tests passed"
fi

#############################################
# Step 2: Build Binaries
#############################################
log_header "Step 2: Building Binaries"

LDFLAGS="-s -w -X main.Version=$VERSION -X main.GitCommit=$GIT_COMMIT -X main.BuildDate=$BUILD_DATE"

if [[ "$DRY_RUN" == "true" ]]; then
    log_info "[DRY RUN] Would build binaries with version $VERSION"
else
    GO=$(command -v go 2>/dev/null || echo /usr/local/go/bin/go)
    
    log_info "Building linux-amd64..."
    GOOS=linux GOARCH=amd64 $GO build -ldflags "$LDFLAGS" -o helm-hooks-linux-amd64 ./cmd/helm-hooks
    
    log_info "Building linux-arm64..."
    GOOS=linux GOARCH=arm64 $GO build -ldflags "$LDFLAGS" -o helm-hooks-linux-arm64 ./cmd/helm-hooks
    
    log_info "Building darwin-amd64..."
    GOOS=darwin GOARCH=amd64 $GO build -ldflags "$LDFLAGS" -o helm-hooks-darwin-amd64 ./cmd/helm-hooks
    
    log_info "Building darwin-arm64..."
    GOOS=darwin GOARCH=arm64 $GO build -ldflags "$LDFLAGS" -o helm-hooks-darwin-arm64 ./cmd/helm-hooks
    
    log_info "Building windows-amd64..."
    GOOS=windows GOARCH=amd64 $GO build -ldflags "$LDFLAGS" -o helm-hooks-windows-amd64.exe ./cmd/helm-hooks
    
    # Also build for current platform
    $GO build -ldflags "$LDFLAGS" -o helm-hooks ./cmd/helm-hooks
    
    log_success "Built all binaries"
    ./helm-hooks version
fi

#############################################
# Step 3: Update Plugin Submodule
#############################################
log_header "Step 3: Updating Plugin Submodule"

if [[ "$DRY_RUN" == "true" ]]; then
    log_info "[DRY RUN] Would copy binary to $PLUGIN_DIR/bin/"
    log_info "[DRY RUN] Would update plugin.yaml version to $VERSION"
else
    # Copy binary to plugin
    mkdir -p "$PLUGIN_DIR/bin"
    cp helm-hooks-linux-amd64 "$PLUGIN_DIR/bin/helm-hooks"
    chmod +x "$PLUGIN_DIR/bin/helm-hooks"
    log_success "Copied binary to plugin"
    
    # Update plugin.yaml version
    sed -i "s/^version:.*/version: $VERSION/" "$PLUGIN_DIR/plugin.yaml"
    log_success "Updated plugin.yaml version"
    
    # Commit plugin submodule
    cd "$PLUGIN_DIR"
    git add .
    git commit -m "Release $TAG" || log_info "No changes to commit in plugin"
    git tag -a "$TAG" -m "Release $TAG" 2>/dev/null || log_info "Tag $TAG already exists in plugin"
    cd "$PROJECT_DIR"
    log_success "Committed plugin submodule"
fi

#############################################
# Step 4: Generate Release Notes
#############################################
log_header "Step 4: Generating Release Notes"

RELEASE_NOTES_DIR="releases"
RELEASE_NOTES_FILE="$RELEASE_NOTES_DIR/$TAG.md"
mkdir -p "$RELEASE_NOTES_DIR"

# Get recent commits
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [[ -n "$LAST_TAG" ]]; then
    COMMITS=$(git log --oneline "$LAST_TAG"..HEAD 2>/dev/null || git log --oneline -10)
else
    COMMITS=$(git log --oneline -10)
fi

cat > "$RELEASE_NOTES_FILE" << EOF
# helm-hooks $TAG

## Release Summary

<!-- TODO: Edit this section with a comprehensive summary -->
[Your release summary here]

---

## Installation

### Helm 4 (Plugin)

\`\`\`bash
helm plugin install https://github.com/$PLUGIN_REPO
\`\`\`

### Helm 3 (Binary)

\`\`\`bash
# Linux (amd64)
curl -sSL https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-linux-amd64 -o helm-hooks
chmod +x helm-hooks

# macOS (arm64)
curl -sSL https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-darwin-arm64 -o helm-hooks
chmod +x helm-hooks
\`\`\`

---

## Downloads

| Platform | Binary |
|----------|--------|
| Linux (amd64) | [helm-hooks-linux-amd64](https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-linux-amd64) |
| Linux (arm64) | [helm-hooks-linux-arm64](https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-linux-arm64) |
| macOS (amd64) | [helm-hooks-darwin-amd64](https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-darwin-amd64) |
| macOS (arm64) | [helm-hooks-darwin-arm64](https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-darwin-arm64) |
| Windows | [helm-hooks-windows-amd64.exe](https://github.com/$GITHUB_REPO/releases/download/$TAG/helm-hooks-windows-amd64.exe) |

---

## Commits

\`\`\`
$COMMITS
\`\`\`

---

**Full Changelog:** https://github.com/$GITHUB_REPO/compare/${LAST_TAG:-main}...$TAG
EOF

log_success "Generated: $RELEASE_NOTES_FILE"

#############################################
# Step 5: Commit and Tag Main Repo
#############################################
log_header "Step 5: Committing Main Repo"

if [[ "$DRY_RUN" == "true" ]]; then
    log_info "[DRY RUN] Would commit and tag $TAG"
else
    git add "$PLUGIN_DIR" "$RELEASE_NOTES_FILE"
    git commit -m "Release $TAG" || log_info "No changes to commit"
    git tag -a "$TAG" -m "Release $TAG"
    log_success "Created tag $TAG"
fi

#############################################
# Step 6: Push to GitHub
#############################################
if [[ "$SKIP_PUSH" == "false" ]]; then
    log_header "Step 6: Pushing to GitHub"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would push plugin and main repo"
    else
        # Push plugin submodule first
        log_info "Pushing plugin submodule..."
        cd "$PLUGIN_DIR"
        git push origin main 2>/dev/null || git push -u origin main
        git push origin "$TAG"
        cd "$PROJECT_DIR"
        log_success "Pushed plugin"
        
        # Push main repo
        log_info "Pushing main repo..."
        git push origin main 2>/dev/null || git push -u origin main
        git push origin "$TAG"
        log_success "Pushed main repo"
    fi
fi

#############################################
# Summary
#############################################
log_header "Release Complete!"

echo "Version: $TAG"
echo
echo "Artifacts:"
echo "  - Binaries: helm-hooks-{linux,darwin,windows}-{amd64,arm64}"
echo "  - Plugin: https://github.com/$PLUGIN_REPO"
echo "  - Release Notes: $RELEASE_NOTES_FILE"
echo
echo "Next steps:"
echo "  1. Edit $RELEASE_NOTES_FILE (add release summary)"
echo "  2. Create GitHub release: https://github.com/$GITHUB_REPO/releases/new?tag=$TAG"
echo "  3. Upload binaries to the GitHub release"
