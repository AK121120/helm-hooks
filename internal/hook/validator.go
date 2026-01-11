package hook

import (
	"fmt"
	"strings"
)

// Valid Helm hook events
var validHooks = map[string]bool{
	"pre-install":         true,
	"post-install":        true,
	"pre-delete":          true,
	"post-delete":         true,
	"pre-upgrade":         true,
	"post-upgrade":        true,
	"pre-rollback":        true,
	"post-rollback":       true,
	"test":                true,
	"test-success":        true, // deprecated but still valid
	"test-failure":        true, // deprecated but still valid
}

// validateHooks validates hook events for a resource.
func validateHooks(hooks []string, resourceName string) error {
	seen := make(map[string]bool)

	for _, h := range hooks {
		// Check for valid hook name
		if !validHooks[h] {
			return fmt.Errorf("resource %q has invalid hook %q", resourceName, h)
		}

		// Check for duplicates
		if seen[h] {
			return fmt.Errorf("resource %q has duplicate hook %q", resourceName, h)
		}
		seen[h] = true
	}

	return nil
}

// ValidateAnnotations performs validation on hook annotations before processing.
func ValidateAnnotations(annotations map[string]string, resourceName string) error {
	hookValue, hasHook := annotations[annotationHook]
	if !hasHook {
		return nil // Not a hook, no validation needed
	}

	hooks := parseHookEvents(hookValue)
	if len(hooks) == 0 {
		return fmt.Errorf("resource %q has empty %s annotation", resourceName, annotationHook)
	}

	// Validate hook names
	if err := validateHooks(hooks, resourceName); err != nil {
		return err
	}

	// Validate weight format if present
	if weightsVal, ok := annotations[annotationHookWeights]; ok {
		if _, err := parseExplicitWeights(weightsVal, hooks); err != nil {
			return fmt.Errorf("resource %q: %w", resourceName, err)
		}
	} else if weightVal, ok := annotations[annotationHookWeight]; ok {
		weightParts := strings.Split(weightVal, ",")
		if len(weightParts) > 1 && len(weightParts) != len(hooks) {
			return fmt.Errorf("resource %q: hook count (%d) does not match weight count (%d)",
				resourceName, len(hooks), len(weightParts))
		}
	}

	return nil
}
