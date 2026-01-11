# helm-hooks

[![CI](https://github.com/AK121120/helm-hooks/actions/workflows/ci.yml/badge.svg)](https://github.com/AK121120/helm-hooks/actions/workflows/ci.yml)
[![Release](https://github.com/AK121120/helm-hooks/actions/workflows/release.yml/badge.svg)](https://github.com/AK121120/helm-hooks/actions/workflows/release.yml)

helm-hooks is an **Advanced Hook Manager** for Helm. It powers up your Helm charts by enabling **per-hook weights**, **auto-splitting** of multi-hook resources, and **environment variable injection**.

Basically, it fixes the limitation where Helm applies one weight to all hooks in a resource.

## The Problem

Helm's `helm.sh/hook-weight` applies the same weight to ALL hooks in a resource. When you need different weights for different hook events (e.g., `pre-install=-100, post-install=200`), you must duplicate the entire resource.

**helm-hooks solves this** by:
- Splitting multi-hook resources into separate ones with individual weights
- Injecting hook context via environment variables
- Supporting multiple weight formats for flexibility

## Quick Start

### Helm 4 (Plugin)

```bash
helm plugin install https://github.com/AK121120/helm-hooks-plugin.git --verify=false
helm install myapp ./chart --post-renderer helm-hooks
```

### Helm 3 (Binary)

```bash
curl -L https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-linux-amd64 -o helm-hooks
chmod +x helm-hooks
helm install myapp ./chart --post-renderer ./helm-hooks
```

## Example

**Input (one resource):**
```yaml
metadata:
  annotations:
    helm.sh/hook: pre-install,post-install
    helm.sh/hook-weights: "pre-install=-100,post-install=200"
```

**Output (two resources with individual weights):**
```yaml
# Resource 1
metadata:
  name: myapp-hook-pre-install
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "-100"
---
# Resource 2
metadata:
  name: myapp-hook-post-install
  annotations:
    helm.sh/hook: post-install
    helm.sh/hook-weight: "200"
```

## Documentation

| Document | Description |
|----------|-------------|
| [Installation](docs/installation.md) | Helm 3 vs Helm 4 setup |
| [Annotations](docs/annotations.md) | All supported annotations |
| [Examples](docs/examples.md) | Usage examples and demo chart |
| [Design](docs/design.md) | Architecture and design decisions |
| [Contributing](docs/contributing.md) | How to contribute |
| [Building](docs/building.md) | Build and release process |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and FAQ |

## Supported Annotations

| Annotation | Description |
|------------|-------------|
| `helm.sh/hook-weights` | Per-hook weight mapping (explicit or positional) |
| `helm.sh/hook-env` | Enable/disable env var injection (`true`/`false`) |
| `helm.sh/hook-name-suffix` | Enable/disable hook name suffix (`true`/`false`) |

## License

MIT
