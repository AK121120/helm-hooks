# Contributing

We welcome contributions!

## Quick Start

1. **Clone**: `git clone --recursive https://github.com/AK121120/helm-hooks.git`
2. **Setup**: Run `make build` to compile.
3. **Test**: Run `make test-all` for unit and E2E validation.

## Development
- **Code Style**: Run `go fmt` before committing.
- **Tests**: Add tests for any new logic.
- **PRs**: Submit PRs to the `main` branch with a clear description.

## Release
Maintainers use `./scripts/release.sh -v X.Y.Z` to automate:
- Testing & Building
- Binary generation
- Plugin updates & git tagging
