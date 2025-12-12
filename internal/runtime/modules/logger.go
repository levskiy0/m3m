package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/pkg/schema"
)

// LogCallback is called when new log is written
type LogCallback func()

type LoggerModule struct {
	file        *os.File
	mu          sync.Mutex
	onLogFunc   LogCallback
	notifyTimer *time.Timer
}

func NewLoggerModule(logPath string) *LoggerModule {
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to stdout if can't create file
		return &LoggerModule{file: os.Stdout}
	}
	return &LoggerModule{file: file}
}

// Name returns the module name for JavaScript
func (l *LoggerModule) Name() string {
	return "$logger"
}

// Register registers the module into the JavaScript VM
func (l *LoggerModule) Register(vm interface{}) {
	v := vm.(*goja.Runtime)
	methods := map[string]interface{}{
		"debug": l.Debug,
		"info":  l.Info,
		"warn":  l.Warn,
		"error": l.Error,
	}
	v.Set(l.Name(), methods)

	// Also register console as an alias
	v.Set("console", map[string]interface{}{
		"log":   l.Info,
		"info":  l.Info,
		"warn":  l.Warn,
		"error": l.Error,
		"debug": l.Debug,
	})
}

// SetOnLog sets a callback that is called when new logs are written
func (l *LoggerModule) SetOnLog(callback LogCallback) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onLogFunc = callback
}

func (l *LoggerModule) log(level string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	message := formatArgs(args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	l.file.WriteString(logLine)
	l.file.Sync() // Flush buffer to disk immediately

	if l.onLogFunc != nil {
		if l.notifyTimer != nil {
			l.notifyTimer.Stop()
		}
		l.notifyTimer = time.AfterFunc(1*time.Second, l.onLogFunc)
	}
}

// formatArgs formats arguments for logging, converting structs/maps/slices to JSON
func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatValue(arg)
	}

	if len(parts) == 1 {
		return parts[0]
	}

	// Join multiple arguments with space
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " " + parts[i]
	}
	return result
}

// formatValue formats a single value, using JSON for complex types
func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	// Handle strings directly
	if s, ok := v.(string); ok {
		return s
	}

	// Check if it's a complex type that should be JSON-serialized
	rv := reflect.ValueOf(v)
	kind := rv.Kind()

	// Dereference pointers
	if kind == reflect.Ptr {
		if rv.IsNil() {
			return "null"
		}
		rv = rv.Elem()
		kind = rv.Kind()
	}

	// Serialize structs, maps, and slices as JSON
	if kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
		jsonBytes, err := json.Marshal(v)
		if err == nil {
			return string(jsonBytes)
		}
	}

	// Fall back to default formatting
	return fmt.Sprint(v)
}

func (l *LoggerModule) Debug(args ...interface{}) {
	l.log("DEBUG", args...)
}

func (l *LoggerModule) Info(args ...interface{}) {
	l.log("INFO", args...)
}

func (l *LoggerModule) Warn(args ...interface{}) {
	l.log("WARN", args...)
}

func (l *LoggerModule) Error(args ...interface{}) {
	l.log("ERROR", args...)
}

func (l *LoggerModule) Close() {
	if l.file != nil && l.file != os.Stdout {
		l.file.Close()
	}
}

// GetSchema implements JSSchemaProvider
func (l *LoggerModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$logger",
		Description: "Logging utilities for debugging and monitoring",
		Methods: []schema.MethodSchema{
			{
				Name:        "debug",
				Description: "Log a debug message",
				Params: []schema.ParamSchema{
					{Name: "args", Type: "any", Description: "Values to log", Variadic: true},
				},
			},
			{
				Name:        "info",
				Description: "Log an info message",
				Params: []schema.ParamSchema{
					{Name: "args", Type: "any", Description: "Values to log", Variadic: true},
				},
			},
			{
				Name:        "warn",
				Description: "Log a warning message",
				Params: []schema.ParamSchema{
					{Name: "args", Type: "any", Description: "Values to log", Variadic: true},
				},
			},
			{
				Name:        "error",
				Description: "Log an error message",
				Params: []schema.ParamSchema{
					{Name: "args", Type: "any", Description: "Values to log", Variadic: true},
				},
			},
		},
	}
}

// GetLoggerSchema returns the logger schema (static version)
func GetLoggerSchema() schema.ModuleSchema {
	return (&LoggerModule{}).GetSchema()
}

// GetConsoleSchema returns schema for console (alias for logger)
func GetConsoleSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "console",
		Description: "Console output (alias for logger)",
		Methods: []schema.MethodSchema{
			{
				Name:        "log",
				Description: "Log a message (alias for info)",
				Params:      []schema.ParamSchema{{Name: "args", Type: "any", Description: "Values to log", Variadic: true}},
			},
			{
				Name:        "info",
				Description: "Log an info message",
				Params:      []schema.ParamSchema{{Name: "args", Type: "any", Description: "Values to log", Variadic: true}},
			},
			{
				Name:        "warn",
				Description: "Log a warning message",
				Params:      []schema.ParamSchema{{Name: "args", Type: "any", Description: "Values to log", Variadic: true}},
			},
			{
				Name:        "error",
				Description: "Log an error message",
				Params:      []schema.ParamSchema{{Name: "args", Type: "any", Description: "Values to log", Variadic: true}},
			},
			{
				Name:        "debug",
				Description: "Log a debug message",
				Params:      []schema.ParamSchema{{Name: "args", Type: "any", Description: "Values to log", Variadic: true}},
			},
		},
	}
}
