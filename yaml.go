package yamlimport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// getPath は、絶対パスの場合はそのまま、相対パスの場合はカレントディレクトリを起点にしたパスを返す関数
func getPath(cdir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(cdir, path)
}

// YAMLファイルを再帰的に処理してimportとプレースホルダを解決する
func processYAML(yamlData map[string]interface{}, cdir string) error {
	if err := resolveImports(yamlData, cdir); err != nil {
		return err
	}

	if err := resolvePlaceholders(yamlData, yamlData); err != nil {
		return err
	}

	return nil
}

// importキーがある場合、ファイルを読み込んでマージする関数
func resolveImports(yamlData map[string]interface{}, cdir string) error {
	for _, value := range yamlData {
		switch v := value.(type) {
		case map[string]interface{}:
			if err := resolveImports(v, cdir); err != nil {
				return err
			}
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if err := resolveImports(itemMap, cdir); err != nil {
						return err
					}
					v[i] = itemMap
				}
			}
		}
	}

	if p, ok := yamlData["import"]; ok {
		path := getPath(cdir, p.(string))
		importData, err := readYAMLFile(path)
		if err != nil {
			return fmt.Errorf("failed to import file: %s err=%v", path, err)
		}
		delete(yamlData, "import")

		for k, v := range importData {
			if _, exists := yamlData[k]; exists {
				return fmt.Errorf("duplicate key: %s", k)
			}
			yamlData[k] = v
		}
	}

	return nil
}

// YAMLファイルを読み込む関数
func readYAMLFile(path string) (map[string]interface{}, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(buf, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// プレースホルダを解決する関数
func resolvePlaceholders(yamlData map[string]interface{}, root map[string]interface{}) error {
	for key, value := range yamlData {
		switch v := value.(type) {
		case string:
			if strings.Contains(v, "{{") && strings.Contains(v, "}}") {
				resolvedValue, err := resolvePlaceholder(v, root)
				if err != nil {
					return err
				}
				yamlData[key] = resolvedValue
			}
		case map[string]interface{}:
			if err := resolvePlaceholders(v, root); err != nil {
				return err
			}
		}
	}
	return nil
}

// プレースホルダを解決するためのサブ関数
func resolvePlaceholder(placeholder string, root map[string]interface{}) (interface{}, error) {
	var result strings.Builder
	startIdx := 0

	for {
		// "{{" の位置を探す
		openIdx := strings.Index(placeholder[startIdx:], "{{")
		if openIdx == -1 {
			// これ以上 "{{" が見つからない場合は、残りの部分をそのまま追加して終了
			result.WriteString(placeholder[startIdx:])
			break
		}

		// "{{" の前の部分を結果に追加
		result.WriteString(placeholder[startIdx : startIdx+openIdx])

		// "}}" の位置を探す
		closeIdx := strings.Index(placeholder[startIdx+openIdx:], "}}")
		if closeIdx == -1 {
			return nil, fmt.Errorf("unmatched '{{' in placeholder: %s", placeholder)
		}

		// "{{" と "}}" の間のキーを取り出す
		key := strings.TrimSpace(placeholder[startIdx+openIdx+2 : startIdx+openIdx+closeIdx])

		// 取り出したキーに基づいて値を取得
		keys := strings.Split(key, ".")
		value, err := getValueFromKeys(root, keys)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve placeholder '%s': %v", key, err)
		}

		// 解決された値を文字列に変換して結果に追加
		result.WriteString(fmt.Sprintf("%v", value))

		// 検索の開始位置を "}}" の後に移動
		startIdx = startIdx + openIdx + closeIdx + 2
	}

	return result.String(), nil
}

// ドットで区切られたキーに対応する値を再帰的に取得する関数
func getValueFromKeys(data map[string]interface{}, keys []string) (interface{}, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("keys are empty")
	}

	value, exists := data[keys[0]]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found", keys[0])
	}

	if len(keys) > 1 {
		nestedMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("key '%s' is not a map", keys[0])
		}
		return getValueFromKeys(nestedMap, keys[1:])
	}

	return value, nil
}

// カスタムUnmarshal関数
func Unmarshal(path string, v interface{}) error {
	// pathのファイルのディレクトリを起点にimportの相対パスを解決する
	cdir := filepath.Dir(path)

	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(buf, &yamlData); err != nil {
		return err
	}

	if err := processYAML(yamlData, cdir); err != nil {
		return err
	}

	buf, err = yaml.Marshal(yamlData)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, v)
}
