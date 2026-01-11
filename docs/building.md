# Building & Releasing

## Prerequisites

- Go 1.22+
- Make
- Git

---

## Build Locally

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run E2E tests
make test-e2e
```

---

## Project Structure

```
helm-hooks/
├── cmd/helm-hooks/         # CLI entry point
│   └── main.go
├── internal/hook/          # Core processing logic
│   ├── processor.go        # Main hook processor
│   ├── splitter.go         # Multi-hook splitter
│   ├── naming.go           # Name generation
│   └── validator.go        # Validation
├── scripts/
│   ├── build.sh            # Build with version injection
│   ├── release.sh          # Automated release script
│   └── test-e2e.sh         # E2E test runner
├── examples/
│   └── demo-chart/         # Demo Helm chart
├── testdata/               # Test input/output files
├── docs/                   # Documentation
├── releases/               # Release notes
└── helm-hooks-plugin/      # Submodule for Helm 4 plugin
```

---

## Release Process

Releases are automated via `scripts/release.sh`:

```bash
# Full release (tests, build, push)
./scripts/release.sh -v 1.2.0

# Dry run (preview only)
./scripts/release.sh -v 1.2.0 --dry-run

# Build only, no push
./scripts/release.sh -v 1.2.0 --skip-push
```

### What the Release Script Does

1. **Run Tests** - Unit and E2E tests
2. **Build Binaries** - All platforms (linux, darwin, windows × amd64, arm64)
3. **Update Plugin** - Copy binary to `helm-hooks-plugin/bin/`
4. **Generate Release Notes** - Create `releases/vX.Y.Z.md`
5. **Commit & Tag** - Both main repo and plugin submodule
6. **Push to GitHub** - Triggers GitHub Actions release workflow

### GitHub Actions

- **CI Workflow** (`ci.yml`): Runs on every push/PR
  - Unit tests
  - E2E tests
  - Vulnerability scan

- **Release Workflow** (`release.yml`): Runs on version tags (`v*`)
  - Builds all binaries
  - Creates GitHub Release
  - Attaches binary artifacts

---

## Version Embedding

Version info is embedded at build time via ldflags:

```bash
go build -ldflags "-X main.Version=1.1.0 -X main.GitCommit=abc123 -X main.BuildDate=2026-01-11" ./cmd/helm-hooks
```

Check version:
```bash
./helm-hooks version
# helm-hooks 1.1.0
#   Git Commit: abc123
#   Build Date: 2026-01-11T12:00:00Z
```

---

## Plugin Submodule

The `helm-hooks-plugin/` directory is a Git submodule pointing to:
https://github.com/AK121120/helm-hooks-plugin

It contains:
- `plugin.yaml` - Helm plugin manifest
- `bin/helm-hooks` - Pre-built binary
- `README.md` - Plugin installation docs

When releasing, the script automatically:
1. Copies the Linux binary to `bin/helm-hooks`
2. Updates `plugin.yaml` version
3. Commits and tags the submodule
4. Pushes to the plugin repository
