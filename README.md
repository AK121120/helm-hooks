# helm-hooks

[![CI](https://github.com/AK121120/helm-hooks/actions/workflows/ci.yml/badge.svg)](https://github.com/AK121120/helm-hooks/actions/workflows/ci.yml)
[![Release](https://github.com/AK121120/helm-hooks/actions/workflows/release.yml/badge.svg)](https://github.com/AK121120/helm-hooks/releases)

A Helm post-renderer that enhances hooks with per-hook weights and environment injection.

## Features

- **Per-hook weights**: Different weights for different hook events
- **Environment injection**: `HELM_HOOK_EVENT` and `HELM_HOOK_WEIGHT`
- **Multi-hook splitting**: Splits resources with multiple hooks
- **Auto-generate hooks**: Derive hooks from `helm.sh/hook-weights`
- **Vulnerability scanned**: All releases pass govulncheck

## Installation

### Helm 4 (Plugin)

```bash
helm plugin install https://github.com/AK121120/helm-hooks-plugin
```

### Helm 3 (Binary)

```bash
# Download latest binary
curl -sSL https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-linux-amd64 -o helm-hooks
chmod +x helm-hooks
sudo mv helm-hooks /usr/local/bin/
```

## Usage

```bash
helm install myapp ./chart --post-renderer helm-hooks
helm upgrade myapp ./chart --post-renderer helm-hooks
```

## Annotations

| Annotation | Description |
|------------|-------------|
| `helm.sh/hook` | Hook events (comma-separated) |
| `helm.sh/hook-weight` | Weight(s) for hooks |
| `helm.sh/hook-weights` | Explicit per-hook weights |
| `helm.sh/hook-env` | Enable/disable env injection (default: `true`) |

## Examples

### Multi-hook with different weights

```yaml
metadata:
  annotations:
    helm.sh/hook: pre-install,post-upgrade
    helm.sh/hook-weights: "pre-install=-10,post-upgrade=100"
```

This creates two separate resources with their respective weights.

### Auto-generate hooks

```yaml
metadata:
  annotations:
    helm.sh/hook-weights: "pre-install=-10,post-delete=100"
```

Hooks are derived from the weights annotation.

## Version

```bash
helm-hooks version
# helm-hooks 1.0.0
#   Git Commit: abc1234
#   Build Date: 2026-01-11T08:00:00Z
```

## Development

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/AK121120/helm-hooks.git

# Build
make build

# Test
make test-all

# Install as Helm 4 plugin (dev mode)
make plugin-install
```

## Release

```bash
# Release a new version
./scripts/release.sh -v 1.0.0

# Preview release (dry run)
./scripts/release.sh -v 1.0.0 --dry-run
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT
