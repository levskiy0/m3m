// Example plugin for M3M
// Build with: go build -buildmode=plugin -o ../example.so
package main

import (
	"fmt"

	"github.com/dop251/goja"
	"m3m/pkg/schema"
)

// ExamplePlugin demonstrates how to create a M3M plugin
type ExamplePlugin struct {
	initialized bool
}

// Name returns the module name for registration in runtime
func (p *ExamplePlugin) Name() string {
	return "example"
}

// Version returns the plugin version
func (p *ExamplePlugin) Version() string {
	return "1.0.0"
}

// Description returns a short description of the plugin
func (p *ExamplePlugin) Description() string {
	return "Example plugin demonstrating M3M plugin system"
}

// Author returns the plugin author name
func (p *ExamplePlugin) Author() string {
	return "M3M Team"
}

// URL returns the plugin homepage or repository URL
func (p *ExamplePlugin) URL() string {
	return "https://github.com/m3m/plugins/example"
}

// Init initializes the plugin with configuration
func (p *ExamplePlugin) Init(config map[string]interface{}) error {
	p.initialized = true
	return nil
}

// RegisterModule registers functions in GOJA runtime
func (p *ExamplePlugin) RegisterModule(runtime *goja.Runtime) error {
	return runtime.Set("example", map[string]interface{}{
		"hello":   p.hello,
		"add":     p.add,
		"reverse": p.reverse,
	})
}

// Shutdown gracefully stops the plugin
func (p *ExamplePlugin) Shutdown() error {
	p.initialized = false
	return nil
}

// GetSchema returns the schema for TypeScript generation
func (p *ExamplePlugin) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "example",
		Description: "Example plugin demonstrating M3M plugin system",
		Methods: []schema.MethodSchema{
			{
				Name:        "hello",
				Description: "Returns a greeting message",
				Params: []schema.ParamSchema{
					{Name: "name", Type: "string", Description: "The name to greet"},
				},
				Returns: &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "add",
				Description: "Adds two numbers",
				Params: []schema.ParamSchema{
					{Name: "a", Type: "number", Description: "First number"},
					{Name: "b", Type: "number", Description: "Second number"},
				},
				Returns: &schema.ParamSchema{Type: "number"},
			},
			{
				Name:        "reverse",
				Description: "Reverses a string",
				Params: []schema.ParamSchema{
					{Name: "str", Type: "string", Description: "String to reverse"},
				},
				Returns: &schema.ParamSchema{Type: "string"},
			},
		},
	}
}

// Plugin functions
func (p *ExamplePlugin) hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func (p *ExamplePlugin) add(a, b float64) float64 {
	return a + b
}

func (p *ExamplePlugin) reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// NewPlugin is the exported function that returns a new plugin instance
// This is required for the plugin loader to work
func NewPlugin() interface{} {
	return &ExamplePlugin{}
}
