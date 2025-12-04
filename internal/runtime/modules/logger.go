package modules

import (
	"fmt"
	"os"
	"sync"
	"time"
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

func (l *LoggerModule) log(level string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprint(args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	l.file.WriteString(logLine)
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
