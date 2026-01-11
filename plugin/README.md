# helm-hooks-plugin

Helm plugin for [helm-hooks](https://github.com/AK121120/helm-hooks).

## Installation

### Helm 4

```bash
helm plugin install https://github.com/AK121120/helm-hooks-plugin
```

### Helm 3

For Helm 3, download the binary directly:

```bash
curl -sSL https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-linux-amd64 -o helm-hooks
chmod +x helm-hooks
```

## Usage

```bash
helm install myapp ./chart --post-renderer helm-hooks
helm upgrade myapp ./chart --post-renderer helm-hooks
```

## Version

Check current version:

```bash
helm-hooks version
```

## Links

- **Main Repository**: https://github.com/AK121120/helm-hooks
- **Container Image**: `quay.io/gkananthakrishna/helm-hooks`
- **Documentation**: https://github.com/AK121120/helm-hooks#readme
