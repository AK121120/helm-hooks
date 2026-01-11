package hook

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	// maxNameLength is the Kubernetes name limit
	maxNameLength = 63
	// hashLength is the length of truncation hash suffix
	hashLength = 8
)

// GenerateName creates a hook-specific name from the original name and hook event.
// It handles the 63-character Kubernetes name limit with deterministic hashing.
func GenerateName(originalName, hookEvent string) string {
	// Create the suffixed name
	suffixedName := originalName + "-" + hookEvent

	// If under limit, use as-is
	if len(suffixedName) <= maxNameLength {
		return suffixedName
	}

	// Need to truncate with hash
	return truncateWithHash(originalName, hookEvent)
}

// truncateWithHash creates a truncated name with a deterministic hash suffix.
// Format: <truncated-name>-<hook>-<hash>
func truncateWithHash(originalName, hookEvent string) string {
	// Calculate hash of original full name
	fullName := originalName + "-" + hookEvent
	hash := sha256.Sum256([]byte(fullName))
	hashStr := hex.EncodeToString(hash[:])[:hashLength]

	// Calculate available space for base name
	// Format: <base>-<hook>-<hash>
	hookPart := "-" + hookEvent
	hashPart := "-" + hashStr
	separatorLen := len(hookPart) + len(hashPart)

	availableLen := maxNameLength - separatorLen
	if availableLen < 1 {
		// Hook name too long, just use hash
		availableLen = maxNameLength - len(hashPart) - 1
		if availableLen < 1 {
			availableLen = 1
		}
	}

	// Truncate the original name
	truncatedBase := truncateName(originalName, availableLen)

	// Build final name
	result := truncatedBase + hookPart + hashPart
	
	// Safety check - if still too long, aggressively truncate
	if len(result) > maxNameLength {
		result = result[:maxNameLength]
	}

	// Ensure valid DNS name (no trailing dash)
	result = strings.TrimRight(result, "-")

	return result
}

// truncateName truncates a name to the given length, avoiding trailing dashes.
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}

	truncated := name[:maxLen]
	// Remove trailing dashes
	truncated = strings.TrimRight(truncated, "-")
	return truncated
}
