package plugin

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/dop251/goja"

	"m3m/internal/config"
	"m3m/pkg/schema"
)

// Plugin interface that all plugins must implement
type Plugin interface {
	// Name returns the module name for registration in runtime
	Name() string

	// Version returns the plugin version
	Version() string

	// Init initializes the plugin with configuration
	Init(config map[string]interface{}) error

	// RegisterModule registers functions in GOJA runtime
	RegisterModule(runtime *goja.Runtime) error

	// Shutdown gracefully stops the plugin
	Shutdown() error

	// GetSchema returns the schema for TypeScript generation
	GetSchema() schema.ModuleSchema
}

// Loader handles loading and managing plugins
type Loader struct {
	plugins map[string]Plugin
	config  *config.Config
	logger  *slog.Logger
}

// NewLoader creates a new plugin loader
func NewLoader(cfg *config.Config, logger *slog.Logger) (*Loader, error) {
	loader := &Loader{
		plugins: make(map[string]Plugin),
		config:  cfg,
		logger:  logger,
	}

	// Load plugins on creation - fail if any plugin fails to load
	if err := loader.LoadAll(); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}

	return loader, nil
}

// LoadAll loads all .so plugins from the plugins directory
func (l *Loader) LoadAll() error {
	pluginsPath := l.config.Plugins.Path
	if pluginsPath == "" {
		pluginsPath = "./plugins"
	}

	// Create plugins directory if it doesn't exist
	if err := os.MkdirAll(pluginsPath, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Scan for .so files
	entries, err := os.ReadDir(pluginsPath)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		pluginPath := filepath.Join(pluginsPath, entry.Name())
		if err := l.Load(pluginPath); err != nil {
			l.logger.Error("Failed to load plugin", "path", pluginPath, "error", err)
			return fmt.Errorf("failed to load plugin %s: %w", entry.Name(), err)
		}
	}

	l.logger.Info("Plugins loaded", "count", len(l.plugins))
	return nil
}

// Load loads a single plugin from the given path
func (l *Loader) Load(path string) error {
	l.logger.Info("Loading plugin", "path", path)

	// Open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for NewPlugin function
	newPluginSym, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin does not export NewPlugin: %w", err)
	}

	// Cast to function - plugins return interface{} for compatibility
	newPluginFunc, ok := newPluginSym.(func() interface{})
	if !ok {
		return fmt.Errorf("NewPlugin has wrong signature, expected func() interface{}")
	}

	// Create plugin instance and cast to Plugin interface
	rawPlugin := newPluginFunc()
	pluginInstance, ok := rawPlugin.(Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement Plugin interface")
	}

	// Initialize plugin with config
	// Plugin config can be stored in main config under plugins.config.<name>
	pluginConfig := make(map[string]interface{})
	// TODO: Load plugin-specific config from config file

	if err := pluginInstance.Init(pluginConfig); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Store plugin
	l.plugins[pluginInstance.Name()] = pluginInstance

	l.logger.Info("Plugin loaded successfully",
		"name", pluginInstance.Name(),
		"version", pluginInstance.Version(),
	)

	return nil
}

// RegisterAll registers all loaded plugins to a GOJA runtime
func (l *Loader) RegisterAll(runtime *goja.Runtime) error {
	for name, p := range l.plugins {
		if err := p.RegisterModule(runtime); err != nil {
			return fmt.Errorf("failed to register plugin %s: %w", name, err)
		}
	}
	return nil
}

// GetPlugin returns a plugin by name
func (l *Loader) GetPlugin(name string) (Plugin, bool) {
	p, ok := l.plugins[name]
	return p, ok
}

// GetAllPlugins returns all loaded plugins
func (l *Loader) GetAllPlugins() map[string]Plugin {
	return l.plugins
}

// GetTypeDefinitions returns combined TypeScript definitions from all plugins
func (l *Loader) GetTypeDefinitions() string {
	var defs strings.Builder
	defs.WriteString("// Plugin type definitions\n\n")

	for name, p := range l.plugins {
		defs.WriteString(fmt.Sprintf("// Plugin: %s v%s\n", name, p.Version()))
		s := p.GetSchema()
		defs.WriteString(s.GenerateTypeScript())
		defs.WriteString("\n\n")
	}

	return defs.String()
}

// Shutdown gracefully stops all plugins
func (l *Loader) Shutdown() error {
	var lastErr error
	for name, p := range l.plugins {
		if err := p.Shutdown(); err != nil {
			l.logger.Error("Failed to shutdown plugin", "name", name, "error", err)
			lastErr = err
		}
	}
	return lastErr
}

// PluginInfo represents information about a loaded plugin
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// GetPluginInfos returns information about all loaded plugins
func (l *Loader) GetPluginInfos() []PluginInfo {
	infos := make([]PluginInfo, 0)
	for _, p := range l.plugins {
		infos = append(infos, PluginInfo{
			Name:    p.Name(),
			Version: p.Version(),
		})
	}
	return infos
}
