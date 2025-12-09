// Package plugin provides the plugin interface and utilities for M3M plugins.
// This package is designed to be imported by external plugins.
package plugin

import (
	"github.com/dop251/goja"

	"github.com/levskiy0/m3m/pkg/schema"
)

// Plugin interface that all plugins must implement
type Plugin interface {
	// Name returns the module name for registration in runtime
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a short description of the plugin
	Description() string

	// Author returns the plugin author name
	Author() string

	// URL returns the plugin homepage or repository URL
	URL() string

	// Init initializes the plugin with configuration
	Init(config map[string]interface{}) error

	// RegisterModule registers functions in GOJA runtime
	RegisterModule(runtime *goja.Runtime) error

	// Shutdown gracefully stops the plugin
	Shutdown() error

	// GetSchema returns the schema for TypeScript generation
	GetSchema() schema.ModuleSchema
}

// PluginInfo represents information about a loaded plugin
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	URL         string `json:"url,omitempty"`
}

// NewPluginFunc is the signature for the NewPlugin function that plugins must export
type NewPluginFunc func() interface{}
