package hook

import (
	"gopkg.in/yaml.v3"
)

// splitResource creates separate resources for each hook event.
func splitResource(node *yaml.Node, res *Resource, hooks []string, weights map[string]int, envEnabled, nameSuffixEnabled bool) ([][]byte, error) {
	var results [][]byte

	for _, hookEvent := range hooks {
		// Deep clone the node
		cloned, err := cloneNode(node)
		if err != nil {
			return nil, err
		}

		// Generate new name
		newName := res.Name
		if nameSuffixEnabled {
			newName = GenerateName(res.Name, hookEvent)
		}

		// Update the cloned resource
		if err := updateSplitResource(cloned, hookEvent, weights[hookEvent], newName, envEnabled); err != nil {
			return nil, err
		}

		// Marshal the result
		out, err := marshalNode(cloned)
		if err != nil {
			return nil, err
		}
		results = append(results, out)
	}

	return results, nil
}

// cloneNode creates a deep copy of a YAML node.
func cloneNode(node *yaml.Node) (*yaml.Node, error) {
	// Marshal and unmarshal to create a deep copy
	data, err := marshalNode(node)
	if err != nil {
		return nil, err
	}

	var cloned yaml.Node
	if err := yaml.Unmarshal(data, &cloned); err != nil {
		return nil, err
	}

	return &cloned, nil
}

// updateSplitResource updates a cloned resource for a specific hook.
func updateSplitResource(node *yaml.Node, hookEvent string, weight int, newName string, envEnabled bool) error {
	content := node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		content = node.Content[0]
	}

	// Update metadata name
	if err := setMetadataName(content, newName); err != nil {
		return err
	}

	// Set single hook event
	if err := setAnnotation(content, annotationHook, hookEvent); err != nil {
		return err
	}

	// Set weight
	if err := setAnnotation(content, annotationHookWeight, itoa(weight)); err != nil {
		return err
	}

	// Remove hook-weights as it's been processed
	removeAnnotation(content, annotationHookWeights)

	// Inject environment variables if enabled
	if envEnabled {
		if err := injectEnvVars(content, hookEvent, weight); err != nil {
			return err
		}
	}

	return nil
}

// setMetadataName updates the name in metadata.
func setMetadataName(node *yaml.Node, name string) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "metadata" {
			metadata := node.Content[i+1]
			return setNameInMetadata(metadata, name)
		}
	}

	return nil
}

// setNameInMetadata sets the name field in metadata.
func setNameInMetadata(metadata *yaml.Node, name string) error {
	if metadata.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(metadata.Content); i += 2 {
		if metadata.Content[i].Value == "name" {
			metadata.Content[i+1].Value = name
			return nil
		}
	}

	return nil
}

// itoa converts int to string (avoiding import)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
