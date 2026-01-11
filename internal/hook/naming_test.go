package hook

import (
	"testing"
)

func TestGenerateName_Short(t *testing.T) {
	name := GenerateName("myapp", "pre-install")
	expected := "myapp-pre-install"
	if name != expected {
		t.Errorf("Expected %q, got %q", expected, name)
	}
}

func TestGenerateName_AtLimit(t *testing.T) {
	// Name that would be exactly 63 chars
	baseName := "a"
	for len(baseName)+len("-pre-install") < 63 {
		baseName += "a"
	}
	
	name := GenerateName(baseName, "pre-install")
	if len(name) > 63 {
		t.Errorf("Name exceeds 63 chars: %d", len(name))
	}
}

func TestGenerateName_Truncation(t *testing.T) {
	// Very long name that requires truncation
	baseName := "this-is-a-very-long-name-that-will-definitely-exceed-the-kubernetes-name-limit"
	name := GenerateName(baseName, "pre-install")
	
	if len(name) > 63 {
		t.Errorf("Name exceeds 63 chars after truncation: %d (%s)", len(name), name)
	}
	
	// Should contain hash for determinism
	if len(name) < 8 {
		t.Error("Name too short, expected hash suffix")
	}
}

func TestGenerateName_Deterministic(t *testing.T) {
	baseName := "myapp-with-a-really-long-name-that-requires-truncation-for-k8s"
	
	name1 := GenerateName(baseName, "pre-install")
	name2 := GenerateName(baseName, "pre-install")
	
	if name1 != name2 {
		t.Errorf("Names should be deterministic: %q vs %q", name1, name2)
	}
}

func TestGenerateName_DifferentHooks(t *testing.T) {
	baseName := "myapp"
	
	preInstall := GenerateName(baseName, "pre-install")
	postInstall := GenerateName(baseName, "post-install")
	
	if preInstall == postInstall {
		t.Error("Different hooks should produce different names")
	}
}
