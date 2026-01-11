# Troubleshooting

Common issues and solutions when using helm-hooks.

## Installation Issues

### Plugin not found (Helm 4)

**Error:** `plugin: "helm-hooks" not found` or `fork/exec bin/helm-hooks: no such file or directory`

**Solution:**
1. Uninstall the plugin: `helm plugin uninstall helm-hooks`
2. Clear plugin cache: `rm -rf ~/.local/share/helm/plugins/helm-hooks*`
3. Reinstall with git: `helm plugin install https://github.com/AK121120/helm-hooks-plugin.git --verify=false`

### Binary permission denied

**Error:** `bash: .../bin/helm-hooks: Permission denied`

**Solution:**
The binary might have lost execute permissions. Run:
```bash
chmod +x $(helm env HELM_PLUGINS)/helm-hooks/bin/helm-hooks
```

---

## Processing Issues

### "did not find expected key" (YAML Error)

**Error:** `parsing YAML: yaml: line X: did not find expected key`

**Cause:** You likely used an invalid format for `hook-weights`.

**Solution:**
Ensure you are using quotes for weights, especially negative ones.

**Wrong:**
```yaml
helm.sh/hook-weights: pre-install=-10,post-install=20
```

**Correct:**
```yaml
helm.sh/hook-weights: "pre-install=-10,post-install=20"
```

### Hooks not splitting / Weights not applied

**Cause:** If you have `helm.sh/hook` but NO `helm.sh/hook-weights` (or single `hook-weight`), helm-hooks operates in **Passthrough Mode** and does NOT modify the resource (except optionally injecting env vars).

**Solution:**
Ensure you have `helm.sh/hook-weights` defined if you want splitting behavior.

### "positional weights count doesn't match"

**Error:** `positional weights count (2) doesn't match hook count (3)`

**Cause:** You used positional format `"-10,20"` but defined 3 hooks.

**Solution:**
Match the number of weights to the number of hooks in `helm.sh/hook`.

---

## Naming Issues

### Multiple resources with same name

**Cause:** You disabled name suffixes (`helm.sh/hook-name-suffix: "false"`) on a multi-hook resource.

**Solution:**
Enable suffixes (default) OR ensure you only have one hook event per resource if you must disable suffixes.
