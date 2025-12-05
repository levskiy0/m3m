package modules

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type LoggerModule struct {
	file *os.File
	mu   sync.Mutex
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

func (l *LoggerModule) log(level string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprint(args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	l.file.WriteString(logLine)
	l.file.Sync() // Flush buffer to disk immediately
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
func (l *LoggerModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "$logger",
		Description: "Logging utilities for debugging and monitoring",
		Methods: []JSMethodSchema{
			{
				Name:        "debug",
				Description: "Log a debug message",
				Params: []JSParamSchema{
					{Name: "args", Type: "any", Description: "Values to log"},
				},
			},
			{
				Name:        "info",
				Description: "Log an info message",
				Params: []JSParamSchema{
					{Name: "args", Type: "any", Description: "Values to log"},
				},
			},
			{
				Name:        "warn",
				Description: "Log a warning message",
				Params: []JSParamSchema{
					{Name: "args", Type: "any", Description: "Values to log"},
				},
			},
			{
				Name:        "error",
				Description: "Log an error message",
				Params: []JSParamSchema{
					{Name: "args", Type: "any", Description: "Values to log"},
				},
			},
		},
	}
}

// GetLoggerSchema returns the logger schema (static version)
func GetLoggerSchema() JSModuleSchema {
	return (&LoggerModule{}).GetSchema()
}

// GetConsoleSchema returns schema for console (alias for logger)
func GetConsoleSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "console",
		Description: "Console output (alias for logger)",
		Methods: []JSMethodSchema{
			{
				Name:        "log",
				Description: "Log a message (alias for info)",
				Params:      []JSParamSchema{{Name: "args", Type: "...any", Description: "Values to log"}},
			},
			{
				Name:        "info",
				Description: "Log an info message",
				Params:      []JSParamSchema{{Name: "args", Type: "...any", Description: "Values to log"}},
			},
			{
				Name:        "warn",
				Description: "Log a warning message",
				Params:      []JSParamSchema{{Name: "args", Type: "...any", Description: "Values to log"}},
			},
			{
				Name:        "error",
				Description: "Log an error message",
				Params:      []JSParamSchema{{Name: "args", Type: "...any", Description: "Values to log"}},
			},
			{
				Name:        "debug",
				Description: "Log a debug message",
				Params:      []JSParamSchema{{Name: "args", Type: "...any", Description: "Values to log"}},
			},
		},
	}
}
