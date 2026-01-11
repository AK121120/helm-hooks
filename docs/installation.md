# Installation

## Helm 4 (Plugin)

Helm 4 supports plugins natively. Install directly from GitHub:

```bash
helm plugin install https://github.com/AK121120/helm-hooks-plugin.git --verify=false
```

**Usage:**
```bash
helm install myapp ./chart --post-renderer helm-hooks
helm upgrade myapp ./chart --post-renderer helm-hooks
```

**Verify installation:**
```bash
helm plugin list
# NAME         VERSION  TYPE              APIVERSION  PROVENANCE  SOURCE
# helm-hooks   1.1.0    postrenderer/v1   v1          unknown     unknown
```

**Uninstall:**
```bash
helm plugin uninstall helm-hooks
```

---

## Helm 3 (Binary)

Helm 3 requires a binary executable for post-rendering.

### Download Binary

```bash
# Linux (amd64)
curl -L https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-linux-amd64 -o helm-hooks
chmod +x helm-hooks

# Linux (arm64)
curl -L https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-linux-arm64 -o helm-hooks
chmod +x helm-hooks

# macOS (arm64 - Apple Silicon)
curl -L https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-darwin-arm64 -o helm-hooks
chmod +x helm-hooks

# macOS (amd64 - Intel)
curl -L https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-darwin-amd64 -o helm-hooks
chmod +x helm-hooks

# Windows
curl -L https://github.com/AK121120/helm-hooks/releases/latest/download/helm-hooks-windows-amd64.exe -o helm-hooks.exe
```

### Usage

```bash
# From current directory
helm install myapp ./chart --post-renderer ./helm-hooks

# Or install to PATH
sudo mv helm-hooks /usr/local/bin/
helm install myapp ./chart --post-renderer helm-hooks
```

### Verify

```bash
./helm-hooks version
# helm-hooks 1.1.0
#   Git Commit: abc123
#   Build Date: 2026-01-11T12:00:00Z
```

---

## Specific Version

To install a specific version:

```bash
# Plugin (Helm 4)
helm plugin install https://github.com/AK121120/helm-hooks-plugin.git --version v1.1.0 --verify=false

# Binary (Helm 3)
curl -L https://github.com/AK121120/helm-hooks/releases/download/v1.1.0/helm-hooks-linux-amd64 -o helm-hooks
chmod +x helm-hooks
```

---

## All Releases

See [GitHub Releases](https://github.com/AK121120/helm-hooks/releases) for all available versions.
