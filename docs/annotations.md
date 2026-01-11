# Supported Annotations

helm-hooks processes and extends the following annotations:

---

## helm.sh/hook-weights

**Purpose:** Specify different weights for each hook event.

### Format 1: Explicit (hook=weight pairs)

```yaml
annotations:
  helm.sh/hook: pre-install,post-install
  helm.sh/hook-weights: "pre-install=-100,post-install=200"
```

### Format 2: Positional (comma-separated weights)

Weights must match the order of hooks:

```yaml
annotations:
  helm.sh/hook: pre-install,post-install
  helm.sh/hook-weights: "-100,200"
```

### Format 3: Auto-generate hooks

If `helm.sh/hook` is absent, hooks are auto-generated from weight keys:

```yaml
annotations:
  # No helm.sh/hook needed!
  helm.sh/hook-weights: "pre-delete=-20,post-delete=20"
```

---

## helm.sh/hook-env

**Purpose:** Control environment variable injection.

**Default:** `true` (enabled)

```yaml
annotations:
  helm.sh/hook-env: "false"  # Disable env var injection
```

When enabled, these env vars are injected into containers:
- `HELM_HOOK_EVENT` - The hook event (e.g., `pre-install`)
- `HELM_HOOK_WEIGHT` - The hook weight (e.g., `-100`)

---

## helm.sh/hook-name-suffix

**Purpose:** Control whether hook name gets a suffix.

**Default:** `true` (enabled)

```yaml
annotations:
  helm.sh/hook-name-suffix: "false"  # Keep original name, no suffix
```

When enabled, resources are renamed with hook suffix:
- `myapp-hook` → `myapp-hook-pre-install`
- `myapp-hook` → `myapp-hook-post-install`

When disabled, all split resources keep the original name.

> [!WARNING]
> **Naming Collisions:** If you disable name suffixes (`"false"`) AND have multiple hooks (e.g., `pre-install,post-install`), helm-hooks will generate multiple resources with the **SAME NAME**. This may cause Helm or Kubernetes to error, or one resource to overwrite the other. Only disable suffixes if you are sure you won't have naming conflicts or are using a single hook event.

---

## Native Helm Annotations (Passthrough)

These native Helm annotations are preserved and NOT modified by helm-hooks:

| Annotation | Description |
|------------|-------------|
| `helm.sh/hook` | Hook events |
| `helm.sh/hook-weight` | Single weight (native Helm) |
| `helm.sh/hook-delete-policy` | Cleanup policy |

**Passthrough behavior:** If a resource has:
- Single hook event with single weight
- No `helm.sh/hook-weights` annotation

Then helm-hooks passes it through unchanged.

---

## Annotation Summary

| Annotation | Type | Default | Description |
|------------|------|---------|-------------|
| `helm.sh/hook-weights` | string | - | Per-hook weights (explicit or positional) |
| `helm.sh/hook-env` | bool | `true` | Inject HELM_HOOK_* env vars |
| `helm.sh/hook-name-suffix` | bool | `true` | Append hook name to resource |
