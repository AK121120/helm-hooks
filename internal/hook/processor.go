// Package hook provides Helm hook enhancement functionality.
// It parses hook annotations, splits multi-hook resources,
// applies per-hook weights, and injects environment variables.
package hook

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// Helm hook annotations
	annotationHook         = "helm.sh/hook"
	annotationHookWeight   = "helm.sh/hook-weight"
	annotationHookWeights  = "helm.sh/hook-weights"
	annotationHookEnv      = "helm.sh/hook-env"
	annotationHookNameSuffix = "helm.sh/hook-name-suffix"
	annotationHookDeletePolicy = "helm.sh/hook-delete-policy"

	// Default weight when not specified
	defaultWeight = 0
)

// Resource represents a Kubernetes resource with typed access to common fields.
type Resource struct {
	node     *yaml.Node
	Kind     string
	Name     string
	Annotations map[string]string
}

// Process takes raw YAML input and returns enhanced YAML output.
// It parses multi-document YAML, processes hooks, and reconstructs the output.
func Process(input []byte) ([]byte, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(input))
	var outputDocs [][]byte

	for {
		var node yaml.Node
		err := decoder.Decode(&node)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("parsing YAML: %w", err)
		}

		// Process this document
		processed, err := processDocument(&node)
		if err != nil {
			return nil, err
		}

		outputDocs = append(outputDocs, processed...)
	}

	// Combine all documents with YAML document separators
	return combineDocuments(outputDocs), nil
}

// processDocument handles a single YAML document.
// Returns one or more YAML byte slices (splitting produces multiple).
func processDocument(node *yaml.Node) ([][]byte, error) {
	// Extract resource metadata
	res, err := parseResource(node)
	if err != nil {
		return nil, err
	}

	// Check for hook annotations
	hookValue, hasHook := res.Annotations[annotationHook]
	weightsValue, hasWeights := res.Annotations[annotationHookWeights]
	weightValue, hasWeight := res.Annotations[annotationHookWeight]

	// Not a hook resource at all - pass through unchanged
	if !hasHook && !hasWeights {
		out, err := marshalNode(node)
		if err != nil {
			return nil, err
		}
		return [][]byte{out}, nil
	}

	var hooks []string

	// Case 1: hook-weights without hook annotation - auto-generate hooks
	if hasWeights && !hasHook {
		hooks, err = extractHooksFromWeights(weightsValue)
		if err != nil {
			return nil, fmt.Errorf("resource %q: %w", res.Name, err)
		}
	} else {
		// Parse hook events from annotation
		hooks = parseHookEvents(hookValue)
		if len(hooks) == 0 {
			return nil, fmt.Errorf("resource %q has empty %s annotation", res.Name, annotationHook)
		}
	}

	// Validate hooks
	if err := validateHooks(hooks, res.Name); err != nil {
		return nil, err
	}

	// Check for passthrough case: single hook with single weight, no hook-weights
	if len(hooks) == 1 && !hasWeights {
		// Single hook - check if we need to modify at all
		if !hasWeight || isSingleValidWeight(weightValue) {
			// Already valid, just add env vars if enabled
			envEnabled := true
			if envVal, ok := res.Annotations[annotationHookEnv]; ok {
				envEnabled = strings.ToLower(envVal) == "true"
			}
			if envEnabled {
				weight := 0
				if hasWeight {
					weight, _ = strconv.Atoi(strings.TrimSpace(weightValue))
				}
				if err := injectEnvVarsOnly(node, hooks[0], weight); err != nil {
					return nil, err
				}
			}
			out, err := marshalNode(node)
			if err != nil {
				return nil, err
			}
			return [][]byte{out}, nil
		}
	}

	// Parse weights for each hook (with validation)
	weights, err := parseWeights(res.Annotations, hooks)
	if err != nil {
		return nil, fmt.Errorf("resource %q: %w", res.Name, err)
	}

	// Check if env injection is enabled (default: true)
	envEnabled := true
	if envVal, ok := res.Annotations[annotationHookEnv]; ok {
		envEnabled = strings.ToLower(envVal) == "true"
	}

	// Check if name suffix is enabled (default: true for multi-hooks)
	nameSuffixEnabled := true
	if suffixVal, ok := res.Annotations[annotationHookNameSuffix]; ok {
		nameSuffixEnabled = strings.ToLower(suffixVal) != "false"
	}

	// Single hook with processing needed
	if len(hooks) == 1 {
		if err := enhanceResource(node, hooks[0], weights[hooks[0]], envEnabled); err != nil {
			return nil, err
		}
		out, err := marshalNode(node)
		if err != nil {
			return nil, err
		}
		return [][]byte{out}, nil
	}

	// Multiple hooks: split into separate resources
	return splitResource(node, res, hooks, weights, envEnabled, nameSuffixEnabled)
}

// extractHooksFromWeights parses hook names from helm.sh/hook-weights
func extractHooksFromWeights(value string) ([]string, error) {
	var hooks []string
	for _, pair := range strings.Split(value, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid weight mapping %q, expected format hook=weight", pair)
		}
		hookName := strings.TrimSpace(parts[0])
		if hookName != "" {
			hooks = append(hooks, hookName)
		}
	}
	if len(hooks) == 0 {
		return nil, fmt.Errorf("no hooks found in hook-weights")
	}
	return hooks, nil
}

// isSingleValidWeight checks if weight value is a single valid number
func isSingleValidWeight(value string) bool {
	value = strings.TrimSpace(value)
	if strings.Contains(value, ",") {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

// injectEnvVarsOnly adds env vars without modifying annotations
func injectEnvVarsOnly(node *yaml.Node, hookEvent string, weight int) error {
	content := node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		content = node.Content[0]
	}
	return injectEnvVars(content, hookEvent, weight)
}

// parseResource extracts metadata from a YAML node.
func parseResource(node *yaml.Node) (*Resource, error) {
	res := &Resource{
		node:        node,
		Annotations: make(map[string]string),
	}

	// Navigate to the document content
	content := node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		content = node.Content[0]
	}

	if content.Kind != yaml.MappingNode {
		return res, nil
	}

	// Extract fields from the mapping
	for i := 0; i < len(content.Content); i += 2 {
		key := content.Content[i]
		value := content.Content[i+1]

		switch key.Value {
		case "kind":
			res.Kind = value.Value
		case "metadata":
			if err := parseMetadata(value, res); err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

// parseMetadata extracts name and annotations from metadata node.
func parseMetadata(node *yaml.Node, res *Resource) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		switch key.Value {
		case "name":
			res.Name = value.Value
		case "annotations":
			if value.Kind == yaml.MappingNode {
				for j := 0; j < len(value.Content); j += 2 {
					annKey := value.Content[j].Value
					annVal := value.Content[j+1].Value
					res.Annotations[annKey] = annVal
				}
			}
		}
	}

	return nil
}

// parseHookEvents splits the hook annotation value into individual events.
func parseHookEvents(value string) []string {
	var hooks []string
	for _, h := range strings.Split(value, ",") {
		h = strings.TrimSpace(h)
		if h != "" {
			hooks = append(hooks, h)
		}
	}
	return hooks
}

// parseWeights determines the weight for each hook event.
// Precedence: helm.sh/hook-weights > comma-separated helm.sh/hook-weight > single weight > default (0)
func parseWeights(annotations map[string]string, hooks []string) (map[string]int, error) {
	weights := make(map[string]int)

	// Initialize all hooks with default weight
	for _, h := range hooks {
		weights[h] = defaultWeight
	}

	// Check for explicit hook-weights mapping (highest priority)
	if weightsVal, ok := annotations[annotationHookWeights]; ok {
		return parseExplicitWeights(weightsVal, hooks)
	}

	// Check for comma-separated hook-weight
	if weightVal, ok := annotations[annotationHookWeight]; ok {
		weightParts := strings.Split(weightVal, ",")

		// Single weight applies to all hooks
		if len(weightParts) == 1 {
			w, err := strconv.Atoi(strings.TrimSpace(weightParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid weight %q: %w", weightVal, err)
			}
			for _, h := range hooks {
				weights[h] = w
			}
			return weights, nil
		}

		// Multiple weights must match hook count
		if len(weightParts) != len(hooks) {
			return nil, fmt.Errorf("hook count (%d) does not match weight count (%d)",
				len(hooks), len(weightParts))
		}

		for i, h := range hooks {
			w, err := strconv.Atoi(strings.TrimSpace(weightParts[i]))
			if err != nil {
				return nil, fmt.Errorf("invalid weight %q: %w", weightParts[i], err)
			}
			weights[h] = w
		}
	}

	return weights, nil
}

// parseExplicitWeights parses helm.sh/hook-weights format: "pre-install=-100,post-install=200"
func parseExplicitWeights(value string, hooks []string) (map[string]int, error) {
	weights := make(map[string]int)

	// Initialize with defaults
	for _, h := range hooks {
		weights[h] = defaultWeight
	}

	// Parse explicit mappings
	for _, pair := range strings.Split(value, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid weight mapping %q, expected format hook=weight", pair)
		}

		hookName := strings.TrimSpace(parts[0])
		weightStr := strings.TrimSpace(parts[1])

		// Verify hook exists
		found := false
		for _, h := range hooks {
			if h == hookName {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("weight specified for unknown hook %q", hookName)
		}

		w, err := strconv.Atoi(weightStr)
		if err != nil {
			return nil, fmt.Errorf("invalid weight for hook %q: %w", hookName, err)
		}
		weights[hookName] = w
	}

	return weights, nil
}

// enhanceResource modifies a resource node to add hook enhancements.
func enhanceResource(node *yaml.Node, hookEvent string, weight int, envEnabled bool) error {
	content := node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		content = node.Content[0]
	}

	// Update the hook-weight annotation
	if err := setAnnotation(content, annotationHookWeight, strconv.Itoa(weight)); err != nil {
		return err
	}

	// Set hook to single event
	if err := setAnnotation(content, annotationHook, hookEvent); err != nil {
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

// setAnnotation sets an annotation value on a resource.
func setAnnotation(node *yaml.Node, key, value string) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	// Find metadata
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "metadata" {
			metadata := node.Content[i+1]
			return setAnnotationInMetadata(metadata, key, value)
		}
	}

	return nil
}

// setAnnotationInMetadata sets an annotation in the metadata node.
func setAnnotationInMetadata(metadata *yaml.Node, key, value string) error {
	if metadata.Kind != yaml.MappingNode {
		return nil
	}

	// Find or create annotations
	for i := 0; i < len(metadata.Content); i += 2 {
		if metadata.Content[i].Value == "annotations" {
			annotations := metadata.Content[i+1]
			return setAnnotationValue(annotations, key, value)
		}
	}

	return nil
}

// setAnnotationValue sets a specific annotation key/value.
// Ensures values are always quoted strings for Kubernetes compatibility.
func setAnnotationValue(annotations *yaml.Node, key, value string) error {
	if annotations.Kind != yaml.MappingNode {
		return nil
	}

	// Find existing key and update
	for i := 0; i < len(annotations.Content); i += 2 {
		if annotations.Content[i].Value == key {
			annotations.Content[i+1].Value = value
			// Force double-quoted string to ensure Kubernetes sees it as a string
			annotations.Content[i+1].Tag = "!!str"
			annotations.Content[i+1].Style = yaml.DoubleQuotedStyle
			return nil
		}
	}

	// Key not found, add it with quoted string value
	annotations.Content = append(annotations.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: "!!str", Style: yaml.DoubleQuotedStyle},
	)

	return nil
}

// removeAnnotation removes an annotation from a resource.
func removeAnnotation(node *yaml.Node, key string) {
	if node.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "metadata" {
			metadata := node.Content[i+1]
			removeAnnotationFromMetadata(metadata, key)
			return
		}
	}
}

// removeAnnotationFromMetadata removes an annotation key.
func removeAnnotationFromMetadata(metadata *yaml.Node, key string) {
	if metadata.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(metadata.Content); i += 2 {
		if metadata.Content[i].Value == "annotations" {
			annotations := metadata.Content[i+1]
			removeAnnotationKey(annotations, key)
			return
		}
	}
}

// removeAnnotationKey removes a specific key from annotations.
func removeAnnotationKey(annotations *yaml.Node, key string) {
	if annotations.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(annotations.Content); i += 2 {
		if annotations.Content[i].Value == key {
			// Remove key and value
			annotations.Content = append(annotations.Content[:i], annotations.Content[i+2:]...)
			return
		}
	}
}

// injectEnvVars adds HELM_HOOK_EVENT and HELM_HOOK_WEIGHT to all containers.
func injectEnvVars(node *yaml.Node, hookEvent string, weight int) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	// Find spec
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "spec" {
			return injectEnvInSpec(node.Content[i+1], hookEvent, weight)
		}
	}

	return nil
}

// injectEnvInSpec handles environment injection in the spec.
func injectEnvInSpec(spec *yaml.Node, hookEvent string, weight int) error {
	if spec.Kind != yaml.MappingNode {
		return nil
	}

	// Handle Job/CronJob nested spec
	for i := 0; i < len(spec.Content); i += 2 {
		key := spec.Content[i].Value
		switch key {
		case "template":
			// Pod template spec
			return injectEnvInPodTemplate(spec.Content[i+1], hookEvent, weight)
		case "jobTemplate":
			// CronJob -> jobTemplate -> spec -> template
			return injectEnvInJobTemplate(spec.Content[i+1], hookEvent, weight)
		case "containers", "initContainers":
			// Direct pod spec
			injectEnvInContainers(spec.Content[i+1], hookEvent, weight)
		}
	}

	return nil
}

// injectEnvInPodTemplate handles pod template spec.
func injectEnvInPodTemplate(template *yaml.Node, hookEvent string, weight int) error {
	if template.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(template.Content); i += 2 {
		if template.Content[i].Value == "spec" {
			return injectEnvInPodSpec(template.Content[i+1], hookEvent, weight)
		}
	}

	return nil
}

// injectEnvInJobTemplate handles CronJob jobTemplate.
func injectEnvInJobTemplate(jobTemplate *yaml.Node, hookEvent string, weight int) error {
	if jobTemplate.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(jobTemplate.Content); i += 2 {
		if jobTemplate.Content[i].Value == "spec" {
			return injectEnvInSpec(jobTemplate.Content[i+1], hookEvent, weight)
		}
	}

	return nil
}

// injectEnvInPodSpec handles container injection in pod spec.
func injectEnvInPodSpec(podSpec *yaml.Node, hookEvent string, weight int) error {
	if podSpec.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(podSpec.Content); i += 2 {
		key := podSpec.Content[i].Value
		if key == "containers" || key == "initContainers" {
			injectEnvInContainers(podSpec.Content[i+1], hookEvent, weight)
		}
	}

	return nil
}

// injectEnvInContainers adds env vars to all containers in a list.
func injectEnvInContainers(containers *yaml.Node, hookEvent string, weight int) {
	if containers.Kind != yaml.SequenceNode {
		return
	}

	for _, container := range containers.Content {
		if container.Kind == yaml.MappingNode {
			injectEnvInContainer(container, hookEvent, weight)
		}
	}
}

// injectEnvInContainer adds env vars to a single container.
func injectEnvInContainer(container *yaml.Node, hookEvent string, weight int) {
	// Find or create env array
	var envNode *yaml.Node
	var envIndex int

	for i := 0; i < len(container.Content); i += 2 {
		if container.Content[i].Value == "env" {
			envNode = container.Content[i+1]
			envIndex = i + 1
			break
		}
	}

	// Create env array if it doesn't exist
	if envNode == nil {
		envNode = &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{}}
		container.Content = append(container.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "env"},
			envNode,
		)
		envIndex = len(container.Content) - 1
	}

	// Add or update HELM_HOOK_EVENT
	addOrUpdateEnvVar(envNode, "HELM_HOOK_EVENT", hookEvent)
	addOrUpdateEnvVar(envNode, "HELM_HOOK_WEIGHT", strconv.Itoa(weight))

	container.Content[envIndex] = envNode
}

// addOrUpdateEnvVar adds or updates an environment variable.
// Ensures values are always quoted strings for Kubernetes compatibility.
func addOrUpdateEnvVar(envNode *yaml.Node, name, value string) {
	if envNode.Kind != yaml.SequenceNode {
		return
	}

	// Check if var already exists
	for _, item := range envNode.Content {
		if item.Kind == yaml.MappingNode {
			for i := 0; i < len(item.Content); i += 2 {
				if item.Content[i].Value == "name" && item.Content[i+1].Value == name {
					// Update existing value
					for j := 0; j < len(item.Content); j += 2 {
						if item.Content[j].Value == "value" {
							item.Content[j+1].Value = value
							item.Content[j+1].Tag = "!!str"
							item.Content[j+1].Style = yaml.DoubleQuotedStyle
							return
						}
					}
					// Add value field if missing
					item.Content = append(item.Content,
						&yaml.Node{Kind: yaml.ScalarNode, Value: "value"},
						&yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: "!!str", Style: yaml.DoubleQuotedStyle},
					)
					return
				}
			}
		}
	}

	// Add new env var with quoted string value
	newEnv := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: name},
			{Kind: yaml.ScalarNode, Value: "value"},
			{Kind: yaml.ScalarNode, Value: value, Tag: "!!str", Style: yaml.DoubleQuotedStyle},
		},
	}
	envNode.Content = append(envNode.Content, newEnv)
}

// marshalNode converts a YAML node back to bytes.
func marshalNode(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// combineDocuments joins multiple YAML documents with separators.
func combineDocuments(docs [][]byte) []byte {
	if len(docs) == 0 {
		return nil
	}

	var result bytes.Buffer
	for i, doc := range docs {
		if i > 0 {
			result.WriteString("---\n")
		}
		result.Write(doc)
	}
	return result.Bytes()
}
