# Design Document

## Overview

`helm-hooks` is a Helm post-renderer that enhances hook functionality without modifying Helm core.

## Architecture

```
stdin (Helm YAML) → helm-hooks → stdout (Enhanced YAML)
```

The binary:
1. Reads multi-document YAML from stdin
2. Identifies hook resources via `helm.sh/hook` annotation
3. Processes enhancements (weights, splitting, env injection)
4. Outputs valid YAML to stdout

## Processing Flow

```
┌──────────────┐     ┌─────────────┐     ┌──────────────┐
│ Parse YAML   │────▶│ Validate    │────▶│ Process Hooks│
│ Documents    │     │ Annotations │     │              │
└──────────────┘     └─────────────┘     └──────┬───────┘
                                                │
                     ┌─────────────────────────────┤
                     │                             │
                     ▼                             ▼
              ┌──────────────┐           ┌──────────────┐
              │ Single Hook  │           │ Multi-Hook   │
              │ (enhance)    │           │ (split)      │
              └──────┬───────┘           └──────┬───────┘
                     │                          │
                     ▼                          ▼
              ┌──────────────────────────────────────┐
              │           Output YAML                │
              └──────────────────────────────────────┘
```

## Weight Precedence

1. `helm.sh/hook-weights` (explicit mapping)
2. `helm.sh/hook-weight` (comma-separated for multi-hook)
3. `helm.sh/hook-weight` (single value applies to all)
4. Default: 0

## Naming Strategy

For multi-hook resources:
- Original name + `-` + hook event
- Example: `myapp-migration` → `myapp-migration-pre-install`

When name would exceed 63 characters:
- Truncate base name
- Append deterministic SHA256 hash (8 chars)
- Example: `very-long-na...-pre-install-a1b2c3d4`

## Environment Injection

Injected into `spec.template.spec.containers[*].env`:
- `HELM_HOOK_EVENT`: The hook phase
- `HELM_HOOK_WEIGHT`: The numeric weight

Handles nested specs (Job, CronJob).

## Error Semantics

Fail fast, never guess:
- Invalid hook name → error
- Duplicate hooks → error  
- Weight parse failure → error
- Hook/weight count mismatch → error

## Non-Goals

- Modifying Helm lifecycle
- Cleaning up historical resources
- Auto-correcting invalid configurations
