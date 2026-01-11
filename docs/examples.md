# Examples

## Demo Chart

The repository includes a demo chart at `examples/demo-chart/` that showcases all supported features.

### Run the Demo

```bash
# Template to see processed output
helm template demo ./examples/demo-chart --post-renderer helm-hooks

# Install to cluster
helm install demo ./examples/demo-chart --post-renderer helm-hooks

# Watch hooks execute
kubectl get pods -l app=demo -w

# Upgrade (triggers post-upgrade hooks)
helm upgrade demo ./examples/demo-chart --post-renderer helm-hooks

# Test (triggers test hook)
helm test demo

# Rollback (triggers rollback hooks)
helm rollback demo 1

# Uninstall (triggers delete hooks)
helm uninstall demo
```

---

## Demo Templates

Each template in the demo chart showcases a different feature:

### 1. pre-hook.yaml - Explicit Weights

```yaml
annotations:
  helm.sh/hook: pre-install,pre-upgrade
  helm.sh/hook-weights: "pre-install=-10,pre-upgrade=-5"
  # helm.sh/hook-env: "true"  (default - env vars injected)
```

**Features:**
- Explicit weight format: `hook=weight`
- Environment variables injected (default)
- Name suffix enabled (default)

### 2. post-hook.yaml - Positional Weights + No Env

```yaml
annotations:
  helm.sh/hook: post-install,post-upgrade
  helm.sh/hook-weights: "10,15"
  helm.sh/hook-env: "false"
```

**Features:**
- Positional weight format: weights match hook order
- Environment injection disabled

### 3. rollback-hook.yaml - Positional + No Suffix

```yaml
annotations:
  helm.sh/hook: pre-rollback,post-rollback
  helm.sh/hook-weights: "-15,25"
  helm.sh/hook-name-suffix: "false"
```

**Features:**
- Positional weights with negative values
- Name suffix disabled (keeps original name)

### 4. delete-hook.yaml - Auto-Generate Hooks

```yaml
annotations:
  # No helm.sh/hook annotation!
  helm.sh/hook-weights: "pre-delete=-20,post-delete=20"
```

**Features:**
- Hooks auto-generated from weight keys
- No `helm.sh/hook` annotation needed

### 5. test-hook.yaml - Native Passthrough

```yaml
annotations:
  helm.sh/hook: test
  helm.sh/hook-weight: "0"
```

**Features:**
- Single hook with native `helm.sh/hook-weight`
- Passed through unchanged (no helm-hooks processing)

---

## Basic Usage Patterns

### Pattern 1: Different weights per hook

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weights: "pre-install=-100,pre-upgrade=-50"
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: myapp:migrate
        command: ["./migrate.sh"]
```

**Output:** Two Jobs with weights -100 and -50.

### Pattern 2: Quick positional weights

```yaml
annotations:
  helm.sh/hook: pre-install,post-install,pre-upgrade,post-upgrade
  helm.sh/hook-weights: "-100,100,-50,50"
```

**Output:** Four Jobs with respective weights.

### Pattern 3: Hook-aware containers

```yaml
containers:
- name: hook
  image: myapp
  command:
  - /bin/sh
  - -c
  - |
    echo "Running $HELM_HOOK_EVENT hook (weight: $HELM_HOOK_WEIGHT)"
    ./run-hook.sh
```

The injected env vars let your scripts know which hook is running.
