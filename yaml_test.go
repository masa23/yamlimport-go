package yamlimport

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getPath(t *testing.T) {
	assert.Equal(t, "/absolute/path", getPath("", "/absolute/path"), "Should return absolute path unchanged")
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, currentDir+"/relative/path", getPath(currentDir, "relative/path"), "Should return correct path for relative path")
}

func Test_resolveImports(t *testing.T) {
	// Importを解決するためのテストファイルを作成
	importYAML := "key1: value1\nkey2: value2\n"
	err := os.WriteFile("test_import.yaml", []byte(importYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("test_import.yaml")

	yamlData := map[string]interface{}{
		"import": "test_import.yaml",
	}
	cdir, err := os.Getwd()
	assert.NoError(t, err)

	err = resolveImports(yamlData, cdir)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}, yamlData, "Should import and merge data correctly")
}

func Test_readYAMLFile(t *testing.T) {
	// テスト用YAMLファイルを作成
	testYAML := "key: value\n"
	err := os.WriteFile("test.yaml", []byte(testYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("test.yaml")

	result, err := readYAMLFile("test.yaml")
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"key": "value"}, result, "Should read YAML file and return correct data")
}

func Test_resolvePlaceholders(t *testing.T) {
	root := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "resolved_value",
		},
	}
	yamlData := map[string]interface{}{
		"placeholder": "{{ outer.inner }}",
	}
	err := resolvePlaceholders(yamlData, root)
	assert.NoError(t, err)
	assert.Equal(t, "resolved_value", yamlData["placeholder"], "Should resolve placeholder to actual value")
}

func Test_processYAML(t *testing.T) {
	// まずインポートされるYAMLファイルを作成します
	importYAML := `
db_host: localhost
db_port: 5432
`
	err := os.WriteFile("import_test.yaml", []byte(importYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("import_test.yaml") // テスト後に削除

	// インポートを含むYAML
	yamlData := map[string]interface{}{
		"import":              "import_test.yaml",
		"app_name":            "TestApp",
		"db_host_placeholder": "{{ db_host }}",
	}

	// processYAML をテスト
	cdir, err := os.Getwd() // カレントディレクトリを取得
	assert.NoError(t, err)
	err = processYAML(yamlData, cdir) // cdirを引数として渡す
	assert.NoError(t, err)

	expected := map[string]interface{}{
		"app_name":            "TestApp",
		"db_host_placeholder": "localhost",
		"db_host":             "localhost",
		"db_port":             5432,
	}

	assert.Equal(t, expected, yamlData)
}

func TestUnmarshal(t *testing.T) {
	// テスト用のインポートされるYAMLファイルを作成します
	importYAML := `
db_host: localhost
db_port: 5432
`
	err := os.WriteFile("import_test.yaml", []byte(importYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("import_test.yaml") // テスト後に削除

	// テスト用のメインYAMLファイルを作成
	mainYAML := `
import: import_test.yaml
app_name: TestApp
db_host_placeholder: "{{ db_host }}"
`
	err = os.WriteFile("main_test.yaml", []byte(mainYAML), 0644)
	assert.NoError(t, err)
	defer os.Remove("main_test.yaml") // テスト後に削除

	// 構造体としてアンマーシャルされるためのマップを定義
	var result map[string]interface{}

	// Unmarshal関数のテスト
	err = Unmarshal("main_test.yaml", &result)
	assert.NoError(t, err)

	// 期待される結果
	expected := map[string]interface{}{
		"app_name":            "TestApp",
		"db_host_placeholder": "localhost",
		"db_host":             "localhost",
		"db_port":             5432,
	}

	assert.Equal(t, expected, result)
}
