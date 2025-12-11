package log

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger struct {
	level LogLevel
	mu    sync.Mutex
}

var defaultLogger *Logger
var once sync.Once

func init() {
	once.Do(func() {
		defaultLogger = &Logger{level: InfoLevel}
	})
}

func SetLevel(level LogLevel) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.level = level
}

func SetLevelFromString(s string) {
	switch s {
	case "debug":
		SetLevel(DebugLevel)
	case "info":
		SetLevel(InfoLevel)
	case "warn":
		SetLevel(WarnLevel)
	case "error":
		SetLevel(ErrorLevel)
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

func (l *Logger) printf(level string, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(format, args...))
	fmt.Fprintln(os.Stdout, msg)
}

func Debug(format string, args ...interface{}) {
	if defaultLogger.shouldLog(DebugLevel) {
		defaultLogger.printf("DEBUG", format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger.shouldLog(InfoLevel) {
		defaultLogger.printf("INFO", format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger.shouldLog(WarnLevel) {
		defaultLogger.printf("WARN", format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger.shouldLog(ErrorLevel) {
		defaultLogger.printf("ERROR", format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	log.Fatalf("[FATAL] "+format, args...)
}
