# yamlimport-go [![Go Report Card](https://goreportcard.com/badge/github.com/masa23/yamlimport-go)](https://goreportcard.com/report/github.com/masa23/yamlimport-go) [![GoDoc](https://godoc.org/github.com/masa23/yamlimport-go?status.svg)](https://godoc.org/github.com/masa23/yamlimport-go) [![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/masa23/yamlimport-go/main/LICENSE)

A YAML import library for Golang. This library supports the functionality of importing other YAML files within a YAML file and also allows dynamic replacement of placeholders (e.g., `{{ Hoge }}`).

## Installation

```bash
go get github.com/masa23/yamlimport-go
```

## Usage

Below is an example of how to use this library to import other files within a YAML file and resolve placeholders.

### Prepare YAML Files

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

### Example Go Program

The following Go program reads `import.yaml`, imports the specified `hoge.yaml`, and resolves the values containing placeholders, storing the data in the struct `Hoge`. Currently, only string type replacements are supported.

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

    // The Unmarshal function directly takes the path of the YAML file and resolves imports and placeholders.
    if err := yamlimport.Unmarshal("import.yaml", &hoge); err != nil {
        log.Fatal(err)
    }

    fmt.Println(hoge.WelcomeMessage) // Output: "Hello, John Doe"
    fmt.Println(hoge.Key1)           // Output: "Value1"
}
```

In this example, the `Unmarshal` function directly takes the file path, resolves the imports, and replaces the placeholders to decode the YAML data into the specified struct.
