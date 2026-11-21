// Package logger is a minimal printf-style global logger.
package logger

import "log"

// Info logs an informational message.
func Info(format string, args ...any) {
	log.Printf("INFO "+format, args...)
}

// Warning logs a warning message.
func Warning(format string, args ...any) {
	log.Printf("WARN "+format, args...)
}

// Error logs an error.
func Error(err error) {
	log.Printf("ERROR %v", err)
}
