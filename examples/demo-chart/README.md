# Demo Helm Chart for helm-hooks

This chart demonstrates the `helm-hooks` post-renderer with all Helm hook types.

## Hooks Included

| Hook Template | Hook Events | Weights |
|--------------|-------------|---------|
| `pre-hook.yaml` | pre-install, pre-upgrade | -10, -5 |
| `post-hook.yaml` | post-install, post-upgrade | 10, 15 |
| `delete-hook.yaml` | pre-delete, post-delete | -20, 20 |
| `rollback-hook.yaml` | pre-rollback, post-rollback | -15, 25 |
| `test-hook.yaml` | test | 0 |

## Setup

Install the plugin (one-time):
```bash
helm plugin install /home/agk/workspace/helm-hooks
```

## Usage

### Preview
```bash
helm template demo ./examples/demo-chart --post-renderer helm-hooks
```

### Install
```bash
helm install demo ./examples/demo-chart --post-renderer helm-hooks
```

### Upgrade
```bash
helm upgrade demo ./examples/demo-chart --post-renderer helm-hooks
```

### Rollback
```bash
helm rollback demo 1 --post-renderer helm-hooks
```

### Test
```bash
helm test demo
```

### Uninstall
```bash
helm uninstall demo
```

## What to Expect

Each hook prints its event and weight:
```
PRE HOOK RUNNING
Hook Event: pre-install
Hook Weight: -10
```

This proves that:
1. Multi-hooks split → `demo-pre-hook-pre-install` + `demo-pre-hook-pre-upgrade`
2. Each has its own weight
3. Environment variables are injected
