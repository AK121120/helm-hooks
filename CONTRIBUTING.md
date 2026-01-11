# Contributing to helm-hooks

Thank you for your interest in contributing!

## Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/AK121120/helm-hooks.git
   cd helm-hooks
   git submodule update --init
   ```

2. **Build:**
   ```bash
   make build
   # or
   ./scripts/build.sh
   ```

3. **Run tests:**
   ```bash
   make test-all
   # or
   ./scripts/test-e2e.sh
   ```

## Making Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make test-all`
5. Commit with clear message: `git commit -m "Add feature X"`
6. Push: `git push origin feature/my-feature`
7. Create a Pull Request

## Code Standards

- Run `go fmt` before committing
- Add tests for new functionality
- Update documentation for user-facing changes
- Follow existing code style

## Release Process

Releases are managed by maintainers using:

```bash
./scripts/release.sh -v MAJOR.MINOR.PATCH
```

This generates release notes with a placeholder for the summary.
The maintainer edits the summary, commits, and re-runs the script.

## Reporting Issues

Please include:
- helm-hooks version (`helm-hooks version`)
- Helm version (`helm version`)
- Kubernetes version (if applicable)
- Minimal reproduction case
