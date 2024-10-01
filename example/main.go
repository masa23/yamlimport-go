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
	NestedKey2     string `yaml:"nested_key2"`
}

func main() {
	var hoge Hoge

	// Unmarshal function directly takes the path of the YAML file and resolves imports and placeholders.
	if err := yamlimport.Unmarshal("test.yaml", &hoge); err != nil {
		log.Fatal(err)
	}

	fmt.Println(hoge.WelcomeMessage) // Output: "Hello, John Doe"
	fmt.Println(hoge.Key1)           // Output: "Value1"
	fmt.Println(hoge.NestedKey2)     // Output: "Value2"
}
