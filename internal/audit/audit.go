package audit

import (
	"fmt"
	"os"
	"time"
)

// Event types
const (
	EventAccess = "ACCESS"
	EventModify = "MODIFY"
	EventDelete = "DELETE"
	EventExport = "EXPORT"
	EventImport = "IMPORT"
)

// Logger handles audit logging
type Logger struct {
	path    string
	enabled bool
}

// New creates a new audit logger
func New(path string) *Logger {
	// Check if audit logging is enabled via environment
	enabled := os.Getenv("PF_AUDIT") != "false"
	
	return &Logger{
		path:    path,
		enabled: enabled,
	}
}

// Log writes an audit event
func (l *Logger) Log(event, key, details string) error {
	if !l.enabled {
		return nil
	}

	// Format log entry
	timestamp := time.Now().Format("2006-01-02T15:04:05Z07:00")
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}
	
	entry := fmt.Sprintf("%s | %s | %s | %s", timestamp, user, event, key)
	if details != "" {
		entry += " | " + details
	}
	entry += "\n"

	// Append to log file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		// Fail silently for audit logs
		return nil
	}
	defer file.Close()

	_, err = file.WriteString(entry)
	return err
}

// SetEnabled enables or disables audit logging
func (l *Logger) SetEnabled(enabled bool) {
	l.enabled = enabled
}