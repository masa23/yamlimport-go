# yamlimport-go [![Go Report Card](https://goreportcard.com/badge/github.com/masa23/yamlimport-go)](https://goreportcard.com/report/github.com/masa23/yamlimport-go) [![GoDoc](https://godoc.org/github.com/masa23/yamlimport-go?status.svg)](https://godoc.org/github.com/masa23/yamlimport-go) [![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/masa23/yamlimport-go/main/LICENSE)

**yamlimport-go** は、Go 向けの軽量な YAML ローダーです。以下の機能をサポートしています：

- `import:` キーによる他の YAML ファイルのインポート（インポート元のファイルからの相対パスで解決）
- マージされた YAML ツリーから `{{ placeholder }}` 形式のプレースホルダを解決
- `yamlimport.Unmarshal(path, &out)` によるシンプルな API

> 注意：現在サポートされているのはスカラ（文字列）の置換のみです。ネストされたオブジェクトや配列のプレースホルダ展開には対応していません。

## インストール

```bash
go get github.com/masa23/yamlimport-go
```

- ファイルのインポート（インポート元の YAML からの相対パス）
- `{{ some.key }}` のようなプレースホルダの置換
- `yamlimport.Unmarshal(path, &out)` によるシンプルな API

> 現在は文字列プレースホルダのみ対応しています。ネストされたリストやマップには未対応です。

## 使用例

以下に、このライブラリを用いて YAML ファイルをインポートし、プレースホルダを解決する方法を示します。

### YAML ファイルを準備

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

### Go プログラムの例

次の Go プログラムでは、`import.yaml` を読み込んで `hoge.yaml` をインポートし、プレースホルダを解決した結果を `Hoge` 構造体に格納します。現在サポートされているのは文字列型の置換のみです。

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

    // Unmarshal 関数は YAML ファイルのパスを受け取り、インポートとプレースホルダの解決を行います。
    if err := yamlimport.Unmarshal("import.yaml", &hoge); err != nil {
        log.Fatal(err)
    }

    fmt.Println(hoge.WelcomeMessage) // 出力: "Hello, John Doe"
    fmt.Println(hoge.Key1)           // 出力: "Value1"
}
```

この例では、`Unmarshal` 関数がファイルパスを直接受け取り、インポートとプレースホルダの解決を行ったうえで、指定した構造体に YAML データをデコードしています。
