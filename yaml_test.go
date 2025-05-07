package yamlimport

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_getPath(t *testing.T) {
	assert.Equal(t, "/absolute/path", getPath("", "/absolute/path"), "Should return absolute path unchanged")
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, currentDir+"/relative/path", getPath(currentDir, "relative/path"), "Should return correct path for relative path")
}

func Test_readYAMLNode(t *testing.T) {
	tempFile := "test_read.yaml"
	content := "key: value\n"
	err := os.WriteFile(tempFile, []byte(content), 0644)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	node, err := readYAMLNode(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, "key", node.Content[0].Content[0].Value)
	assert.Equal(t, "value", node.Content[0].Content[1].Value)
}

func mustYAMLString(n *yaml.Node) string {
	out, err := yaml.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func Test_resolvePlaceholders(t *testing.T) {
	yamlStr := `
name: "App"
env:
  region: ap-northeast-1
  zone: "c"
message: "Deploying to {{ env.region }}-{{ env.zone }}"
nested:
  note: "{{ name }} is live"
`
	var root yaml.Node
	require.NoError(t, yaml.NewDecoder(strings.NewReader(yamlStr)).Decode(&root))
	require.Len(t, root.Content, 1)
	mapping := root.Content[0]

	err := resolvePlaceholders(mapping, mapping)
	require.NoError(t, err)

	var out map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(mustYAMLString(&root)), &out))

	assert.Equal(t, "Deploying to ap-northeast-1-c", out["message"])
	assert.Equal(t, map[string]interface{}{
		"note": "App is live",
	}, out["nested"])
}

func Test_resolvePlaceholder(t *testing.T) {
	yamlData := `
env:
  name: production
  region: ap-northeast-1
service:
  name: api
  port: "8080"
`
	var root yaml.Node
	require.NoError(t, yaml.NewDecoder(strings.NewReader(yamlData)).Decode(&root))

	tests := []struct {
		name     string
		template string
		want     string
		wantErr  bool
	}{
		{
			name:     "Single placeholder",
			template: "Environment: {{ env.name }}",
			want:     "Environment: production",
		},
		{
			name:     "Multiple placeholders",
			template: "{{ service.name }} running in {{ env.region }}",
			want:     "api running in ap-northeast-1",
		},
		{
			name:     "Nested placeholder",
			template: "Port: {{ service.port }}",
			want:     "Port: 8080",
		},
		{
			name:     "Missing key",
			template: "{{ env.zone }}",
			wantErr:  true,
		},
		{
			name:     "Unclosed placeholder",
			template: "Invalid {{ env.name",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePlaceholder(tt.template, root.Content[0])
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	// Create a YAML file to be imported for testing
	importYAML := `
db_host: localhost
db_port: 5432
`
	err := os.WriteFile("import_test.yaml", []byte(importYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("import_test.yaml") // Remove after test

	// Create the main YAML file for testing
	mainYAML := `
app_name: TestApp
db_host_placeholder: "{{ db_host }}"
import: import_test.yaml
`
	err = os.WriteFile("main_test.yaml", []byte(mainYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("main_test.yaml") // Remove after test

	// Define a map for unmarshalling the result
	var result map[string]interface{}

	// Test the Unmarshal function
	err = Unmarshal("main_test.yaml", &result)
	assert.NoError(t, err)

	// Expected result
	expected := map[string]interface{}{
		"app_name":            "TestApp",
		"db_host_placeholder": "localhost",
		"db_host":             "localhost",
		"db_port":             5432,
	}

	assert.Equal(t, expected, result)
}
