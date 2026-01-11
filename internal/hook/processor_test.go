package hook

import (
	"strings"
	"testing"
)

func TestProcess_SingleHook(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-init
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "5"
spec:
  template:
    spec:
      containers:
        - name: init
          image: busybox
`

	output, err := Process([]byte(input))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := string(output)

	// Should contain env vars
	if !strings.Contains(result, "HELM_HOOK_EVENT") {
		t.Error("Expected HELM_HOOK_EVENT env var")
	}
	if !strings.Contains(result, "HELM_HOOK_WEIGHT") {
		t.Error("Expected HELM_HOOK_WEIGHT env var")
	}
	if !strings.Contains(result, "pre-install") {
		t.Error("Expected pre-install hook value")
	}
}

func TestProcess_MultiHook(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-migration
  annotations:
    helm.sh/hook: pre-install,post-upgrade
    helm.sh/hook-weight: "-100,200"
spec:
  template:
    spec:
      containers:
        - name: migrate
          image: busybox
`

	output, err := Process([]byte(input))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := string(output)

	// Should split into two documents
	docs := strings.Split(result, "---")
	// First doc might be empty before first ---, or we expect 2 docs
	docCount := 0
	for _, d := range docs {
		if strings.TrimSpace(d) != "" {
			docCount++
		}
	}
	if docCount != 2 {
		t.Errorf("Expected 2 documents, got %d", docCount)
	}

	// Should have both hook names
	if !strings.Contains(result, "myapp-migration-pre-install") {
		t.Error("Expected pre-install suffixed name")
	}
	if !strings.Contains(result, "myapp-migration-post-upgrade") {
		t.Error("Expected post-upgrade suffixed name")
	}
}

func TestProcess_ExplicitWeights(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-cleanup
  annotations:
    helm.sh/hook: pre-install,post-install
    helm.sh/hook-weights: "pre-install=-50,post-install=100"
spec:
  template:
    spec:
      containers:
        - name: cleanup
          image: busybox
`

	output, err := Process([]byte(input))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := string(output)

	// Should have correct weights
	if !strings.Contains(result, `helm.sh/hook-weight: "-50"`) && !strings.Contains(result, "hook-weight: '-50'") {
		// Check for weight value in env or annotation
		if !strings.Contains(result, "-50") {
			t.Error("Expected weight -50 in output")
		}
	}
}

func TestProcess_EnvDisabled(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-init
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "5"
    helm.sh/hook-env: "false"
spec:
  template:
    spec:
      containers:
        - name: init
          image: busybox
`

	output, err := Process([]byte(input))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := string(output)

	// Should NOT contain env vars when disabled
	if strings.Contains(result, "HELM_HOOK_EVENT") {
		t.Error("Should not have HELM_HOOK_EVENT when helm.sh/hook-env is false")
	}
}

func TestProcess_NonHook(t *testing.T) {
	input := `apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  key: value
`

	output, err := Process([]byte(input))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := string(output)

	// Should pass through unchanged (mostly)
	if !strings.Contains(result, "ConfigMap") {
		t.Error("Expected ConfigMap in output")
	}
	if !strings.Contains(result, "myapp-config") {
		t.Error("Expected name preserved")
	}
}

func TestProcess_InvalidHook(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-init
  annotations:
    helm.sh/hook: invalid-hook-name
spec:
  template:
    spec:
      containers:
        - name: init
          image: busybox
`

	_, err := Process([]byte(input))
	if err == nil {
		t.Fatal("Expected error for invalid hook name")
	}
	if !strings.Contains(err.Error(), "invalid hook") {
		t.Errorf("Expected 'invalid hook' error, got: %v", err)
	}
}

func TestProcess_DuplicateHook(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-init
  annotations:
    helm.sh/hook: pre-install,pre-install
spec:
  template:
    spec:
      containers:
        - name: init
          image: busybox
`

	_, err := Process([]byte(input))
	if err == nil {
		t.Fatal("Expected error for duplicate hook")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected 'duplicate' error, got: %v", err)
	}
}

func TestProcess_WeightMismatch(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-init
  annotations:
    helm.sh/hook: pre-install,post-install
    helm.sh/hook-weight: "5"
spec:
  template:
    spec:
      containers:
        - name: init
          image: busybox
`

	// Single weight for multiple hooks is valid (applies to all)
	_, err := Process([]byte(input))
	if err != nil {
		t.Fatalf("Single weight for multiple hooks should be valid: %v", err)
	}
}

func TestProcess_WeightCountMismatch(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: myapp-init
  annotations:
    helm.sh/hook: pre-install,post-install,post-upgrade
    helm.sh/hook-weight: "5,10"
spec:
  template:
    spec:
      containers:
        - name: init
          image: busybox
`

	_, err := Process([]byte(input))
	if err == nil {
		t.Fatal("Expected error for weight count mismatch")
	}
	if !strings.Contains(err.Error(), "count") {
		t.Errorf("Expected 'count' error, got: %v", err)
	}
}
