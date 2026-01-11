#!/bin/bash
# E2E Regression Tests for helm-hooks
# This script tests all positive and negative scenarios

# Don't use set -e due to bash arithmetic returning non-zero

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/helm-hooks"
TESTDATA="$PROJECT_DIR/testdata"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    echo "  Expected: $2"
    echo "  Got: $3"
    ((FAILED++))
}

log_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Ensure binary exists
if [[ ! -f "$BINARY" ]]; then
    echo "Building binary..."
    cd "$PROJECT_DIR"
    go build -o helm-hooks ./cmd/helm-hooks
fi

echo "========================================"
echo "E2E Regression Tests for helm-hooks"
echo "========================================"
echo

#############################################
# POSITIVE TESTS
#############################################

log_info "POSITIVE TESTS"

# Test 1: Single hook passthrough (should keep original, add env vars)
test_single_hook_passthrough() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "-10"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1)
    
    # Should preserve original weight (quoted)
    if echo "$output" | grep -q 'helm.sh/hook-weight: "-10"'; then
        log_pass "Single hook passthrough - weight preserved"
    else
        log_fail "Single hook passthrough - weight preserved" '"-10"' "$(echo "$output" | grep hook-weight)"
    fi
    
    # Should add env vars
    if echo "$output" | grep -q 'HELM_HOOK_EVENT'; then
        log_pass "Single hook passthrough - env vars injected"
    else
        log_fail "Single hook passthrough - env vars injected" "HELM_HOOK_EVENT" "not found"
    fi
}

# Test 2: Multi-hook splitting with comma weights
test_multi_hook_split() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install,post-upgrade
    helm.sh/hook-weight: "-10,100"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1)
    
    # Should create two documents
    local doc_count
    doc_count=$(echo "$output" | grep -c "^---" || echo "0")
    if [[ "$doc_count" -ge 1 ]]; then
        log_pass "Multi-hook split - creates multiple documents"
    else
        log_fail "Multi-hook split - creates multiple documents" ">=1 separators" "$doc_count"
    fi
    
    # Should have pre-install resource
    if echo "$output" | grep -q 'test-job-pre-install'; then
        log_pass "Multi-hook split - pre-install name"
    else
        log_fail "Multi-hook split - pre-install name" "test-job-pre-install" "not found"
    fi
    
    # Should have post-upgrade resource
    if echo "$output" | grep -q 'test-job-post-upgrade'; then
        log_pass "Multi-hook split - post-upgrade name"
    else
        log_fail "Multi-hook split - post-upgrade name" "test-job-post-upgrade" "not found"
    fi
}

# Test 3: Explicit hook-weights
test_explicit_weights() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install,post-install
    helm.sh/hook-weights: "pre-install=-50,post-install=100"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1)
    
    # Should have correct weight for pre-install
    if echo "$output" | grep -A5 'pre-install' | grep -q '"-50"'; then
        log_pass "Explicit weights - pre-install has -50"
    else
        log_fail "Explicit weights - pre-install has -50" "-50" "not found"
    fi
}

# Test 4: hook-weights without hook annotation (auto-generate)
test_auto_generate_hooks() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook-weights: "pre-install=-10,post-delete=100"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1)
    
    # Should create resources for both hooks
    if echo "$output" | grep -q 'pre-install'; then
        log_pass "Auto-generate hooks - pre-install created"
    else
        log_fail "Auto-generate hooks - pre-install created" "pre-install" "not found"
    fi
    
    if echo "$output" | grep -q 'post-delete'; then
        log_pass "Auto-generate hooks - post-delete created"
    else
        log_fail "Auto-generate hooks - post-delete created" "post-delete" "not found"
    fi
}

# Test 5: Non-hook passthrough
test_non_hook_passthrough() {
    local input='apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1)
    
    if echo "$output" | grep -q 'ConfigMap'; then
        log_pass "Non-hook passthrough - ConfigMap preserved"
    else
        log_fail "Non-hook passthrough - ConfigMap preserved" "ConfigMap" "not found"
    fi
    
    # Should NOT have HELM_HOOK_EVENT
    if ! echo "$output" | grep -q 'HELM_HOOK_EVENT'; then
        log_pass "Non-hook passthrough - no env vars added"
    else
        log_fail "Non-hook passthrough - no env vars added" "no HELM_HOOK_EVENT" "found"
    fi
}

# Test 6: Env injection disabled
test_env_disabled() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-env: "false"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1)
    
    if ! echo "$output" | grep -q 'HELM_HOOK_EVENT'; then
        log_pass "Env disabled - no env vars added"
    else
        log_fail "Env disabled - no env vars added" "no HELM_HOOK_EVENT" "found"
    fi
}

#############################################
# NEGATIVE TESTS
#############################################

echo
log_info "NEGATIVE TESTS"

# Test 7: Invalid hook name
test_invalid_hook() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: invalid-hook-name
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1) || true
    
    if echo "$output" | grep -qi 'invalid hook'; then
        log_pass "Invalid hook - error returned"
    else
        log_fail "Invalid hook - error returned" "invalid hook error" "$output"
    fi
}

# Test 8: Duplicate hooks
test_duplicate_hook() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install,pre-install
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1) || true
    
    if echo "$output" | grep -qi 'duplicate'; then
        log_pass "Duplicate hook - error returned"
    else
        log_fail "Duplicate hook - error returned" "duplicate error" "$output"
    fi
}

# Test 9: Weight count mismatch
test_weight_count_mismatch() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install,post-install,post-upgrade
    helm.sh/hook-weight: "-10,100"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1) || true
    
    if echo "$output" | grep -qi 'count'; then
        log_pass "Weight count mismatch - error returned"
    else
        log_fail "Weight count mismatch - error returned" "count error" "$output"
    fi
}

# Test 10: Explicit weights with unknown hook
test_explicit_weight_unknown_hook() {
    local input='apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
  annotations:
    helm.sh/hook: pre-install,pre-delete
    helm.sh/hook-weights: "pre-install=-10,post-install=100"
spec:
  template:
    spec:
      containers:
        - name: test
          image: busybox'

    local output
    output=$(echo "$input" | "$BINARY" 2>&1) || true
    
    if echo "$output" | grep -qi 'unknown hook'; then
        log_pass "Unknown hook in weights - error returned"
    else
        log_fail "Unknown hook in weights - error returned" "unknown hook error" "$output"
    fi
}

#############################################
# RUN ALL TESTS
#############################################

test_single_hook_passthrough
test_multi_hook_split
test_explicit_weights
test_auto_generate_hooks
test_non_hook_passthrough
test_env_disabled
test_invalid_hook
test_duplicate_hook
test_weight_count_mismatch
test_explicit_weight_unknown_hook

#############################################
# SUMMARY
#############################################

echo
echo "========================================"
echo "RESULTS"
echo "========================================"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo

if [[ $FAILED -gt 0 ]]; then
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
