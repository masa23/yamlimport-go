package yamlimport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// expandImports recursively loads and merges YAML files specified via "import" keys.
func expandImports(n *yaml.Node, cdir string) error {
	if n.Kind == yaml.MappingNode {
		i := 0
		for i < len(n.Content) {
			keyNode := n.Content[i]
			valNode := n.Content[i+1]

			// Detect "import: path.yaml"
			if keyNode.Value == "import" && valNode.Kind == yaml.ScalarNode {
				path := getPath(cdir, valNode.Value)
				imported, err := readYAMLNode(path)
				if err != nil {
					return err
				}
				if imported.Kind == yaml.DocumentNode && len(imported.Content) > 0 {
					imported = imported.Content[0]
				}
				if imported.Kind != yaml.MappingNode {
					return fmt.Errorf("imported file must be a mapping node")
				}
				// Remove "import" key and check for duplicate keys before merging contents
				n.Content = append(n.Content[:i], n.Content[i+2:]...)
				existingKeys := make(map[string]bool)
				for j := 0; j < len(n.Content); j += 2 {
					existingKeys[n.Content[j].Value] = true
				}
				for j := 0; j < len(imported.Content); j += 2 {
					if existingKeys[imported.Content[j].Value] {
						return fmt.Errorf("duplicate key '%s' found during import", imported.Content[j].Value)
					}
				}
				n.Content = append(n.Content, imported.Content...)
				// Restart scan to handle nested imports
				if err := expandImports(n, cdir); err != nil {
					return err
				}
				i = 0
				continue
			}

			// Recursively expand if value is MappingNode or SequenceNode
			if valNode.Kind == yaml.MappingNode {
				if err := expandImports(valNode, cdir); err != nil {
					return err
				}
			} else if valNode.Kind == yaml.SequenceNode {
				for _, item := range valNode.Content {
					if err := expandImports(item, cdir); err != nil {
						return err
					}
				}
			}
			i += 2
		}
	} else if n.Kind == yaml.SequenceNode {
		for _, item := range n.Content {
			if err := expandImports(item, cdir); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolvePlaceholders replaces "{{key}}" placeholders with corresponding values in the YAML node tree.
func resolvePlaceholders(n *yaml.Node, root *yaml.Node) error {
	switch n.Kind {
	case yaml.MappingNode:
		for i := 1; i < len(n.Content); i += 2 {
			if err := resolvePlaceholders(n.Content[i], root); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for _, item := range n.Content {
			if err := resolvePlaceholders(item, root); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		if n.Tag == "!!str" && strings.Contains(n.Value, "{{") {
			resolved, err := resolvePlaceholder(n.Value, root)
			if err != nil {
				return err
			}
			n.Value = resolved
		}
	}
	return nil
}

// getPath resolves relative path to an absolute one based on the current directory.
func getPath(cdir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(cdir, path)
}

// readYAMLNode reads a YAML file and returns its root node.
func readYAMLNode(path string) (*yaml.Node, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(buf, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// resolvePlaceholder replaces a template string containing {{key}} with actual value from root node.
func resolvePlaceholder(template string, root *yaml.Node) (string, error) {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	var result strings.Builder
	startIdx := 0
	for {
		openIdx := strings.Index(template[startIdx:], "{{")
		if openIdx == -1 {
			result.WriteString(template[startIdx:])
			break
		}
		result.WriteString(template[startIdx : startIdx+openIdx])
		closeIdx := strings.Index(template[startIdx+openIdx:], "}}")
		if closeIdx == -1 {
			return "", fmt.Errorf("unmatched '{{' in: %s", template)
		}
		key := strings.TrimSpace(template[startIdx+openIdx+2 : startIdx+openIdx+closeIdx])
		val, err := findValueInNode(root, strings.Split(key, "."))
		if err != nil {
			return "", err
		}
		result.WriteString(val)
		startIdx = startIdx + openIdx + closeIdx + 2
	}
	return result.String(), nil
}

// findValueInNode retrieves a scalar string value by key path (e.g., ["nested", "key"]) from YAML node.
func findValueInNode(n *yaml.Node, keys []string) (string, error) {
	for _, k := range keys {
		if n.Kind != yaml.MappingNode {
			return "", fmt.Errorf("expected map at %s", k)
		}
		found := false
		for i := 0; i < len(n.Content); i += 2 {
			key := n.Content[i]
			val := n.Content[i+1]
			if key.Value == k {
				n = val
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("key not found: %s", k)
		}
	}
	if n.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("expected scalar for placeholder value, got kind %d", n.Kind)
	}
	return n.Value, nil
}

// Unmarshal reads a YAML file, resolves imports and placeholders, and unmarshals into a Go struct.
func Unmarshal(path string, v interface{}) error {
	cdir := filepath.Dir(path)
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(buf, &root); err != nil {
		return err
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("invalid YAML: expected mapping at document root")
	}

	mapping := root.Content[0]

	// 1. Expand imports
	if err := expandImports(mapping, cdir); err != nil {
		return err
	}

	// 2. Resolve placeholders
	if err := resolvePlaceholders(mapping, mapping); err != nil {
		return err
	}

	// Marshal back to bytes and unmarshal into struct
	newBuf, err := yaml.Marshal(&root)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(newBuf, v)
}
