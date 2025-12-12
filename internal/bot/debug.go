// Himiko Discord Bot
// Copyright (C) 2025 Himiko Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package bot

import (
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"strings"
)

// DebugLogger provides debug logging functionality
type DebugLogger struct {
	enabled bool
}

// NewDebugLogger creates a new debug logger
func NewDebugLogger(enabled bool) *DebugLogger {
	return &DebugLogger{enabled: enabled}
}

// Log logs a debug message if debug mode is enabled
func (d *DebugLogger) Log(format string, args ...interface{}) {
	if d.enabled {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// LogError logs an error with optional stack trace
func (d *DebugLogger) LogError(err error, context string) {
	if err == nil {
		return
	}

	if d.enabled {
		log.Printf("[ERROR] %s: %v", context, err)
		log.Printf("[STACK TRACE]\n%s", string(debug.Stack()))
	} else {
		log.Printf("[ERROR] %s: %v", context, err)
	}
}

// LogPanic recovers from a panic and logs the stack trace
func (d *DebugLogger) LogPanic(context string) {
	if r := recover(); r != nil {
		log.Printf("[PANIC] %s: %v", context, r)
		log.Printf("[STACK TRACE]\n%s", string(debug.Stack()))
	}
}

// GetStackTrace returns the current stack trace as a string
func GetStackTrace() string {
	return string(debug.Stack())
}

// GetCallerInfo returns information about the calling function
func GetCallerInfo(skip int) string {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		// Get just the function name without the full path
		if idx := strings.LastIndex(funcName, "/"); idx >= 0 {
			funcName = funcName[idx+1:]
		}
	}

	// Get just the filename without the full path
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s:%d %s", file, line, funcName)
}

// LogWithCaller logs a message with caller information
func (d *DebugLogger) LogWithCaller(format string, args ...interface{}) {
	if d.enabled {
		caller := GetCallerInfo(1)
		log.Printf("[DEBUG] [%s] "+format, append([]interface{}{caller}, args...)...)
	}
}

// PrintMemStats prints memory statistics (useful for debugging memory issues)
func (d *DebugLogger) PrintMemStats() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("[DEBUG] Memory Stats:")
	log.Printf("  Alloc: %d MB", m.Alloc/1024/1024)
	log.Printf("  TotalAlloc: %d MB", m.TotalAlloc/1024/1024)
	log.Printf("  Sys: %d MB", m.Sys/1024/1024)
	log.Printf("  NumGC: %d", m.NumGC)
	log.Printf("  Goroutines: %d", runtime.NumGoroutine())
}
