package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RenderJSON outputs data as formatted JSON
func RenderJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// RenderYAML outputs data as formatted YAML
func RenderYAML(v interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(v)
}

// RenderError outputs an error message to stderr
func RenderError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// RenderSuccess outputs a success message
func RenderSuccess(message string) {
	fmt.Println(message)
}

// RenderWarning outputs a warning message to stderr
func RenderWarning(message string) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", message)
}
