# yamlimport-go [![Go Report Card](https://goreportcard.com/badge/github.com/masa23/yamlimport-go)](https://goreportcard.com/report/github.com/masa23/yamlimport-go) [![GoDoc](https://godoc.org/github.com/masa23/yamlimport-go?status.svg)](https://godoc.org/github.com/masa23/yamlimport-go) [![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/masa23/yamlimport-go/main/LICENSE)

Golang用のYAMLインポートライブラリです。このライブラリを使用することで、YAMLファイル内で他のYAMLファイルをインポートする機能をサポートし、また、プレースホルダー（例：`{{ Hoge }}`）の動的置き換えも可能になります。

## インストール

```bash
go get github.com/masa23/yamlimport-go
```

## 使い方

このライブラリを使用して、YAMLファイル内で他のファイルをインポートし、プレースホルダーを解決する例を以下に示します。

### YAMLファイルの準備

`import.yaml`:
```yaml
import: hoge.yaml
welcome_message: "Hello, {{ UserName }}"
```

`hoge.yaml`:
```yaml
UserName: "John Doe"
Key1: Value1
```

### Goプログラムの例

以下のGoプログラムでは、`import.yaml` を読み込み、その中で指定された `hoge.yaml` をインポートし、プレースホルダーを含む値を解決して、構造体 `Hoge` にデータを格納します。
現時点では、置き換えはstring型のみサポートされています。

```go
package main

import (
    "fmt"
    "log"
    "github.com/masa23/yamlimport-go"
)

type Hoge struct {
    UserName       string `yaml:"UserName"`
    Key1           string `yaml:"Key1"`
    WelcomeMessage string `yaml:"welcome_message"`
}

func main() {
    var hoge Hoge

    // Unmarshal function directly takes the path of the YAML file and resolves imports and placeholders.
    if err := yamlimport.Unmarshal("import.yaml", &hoge); err != nil {
        log.Fatal(err)
    }

    fmt.Println(hoge.WelcomeMessage) // Output: "Hello, John Doe"
    fmt.Println(hoge.Key1)           // Output: "Value1"
}
```

この例では、`Unmarshal` 関数が直接ファイルパスを取り、インポートを解決し、プレースホルダーを置換して指定された構造体にYAMLデータをデコードします。
