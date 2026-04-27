package audit

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Level represents the severity of an audit event.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Level     Level     `json:"level"`
	Action    string    `json:"action"`
	Provider  string    `json:"provider,omitempty"`
	Path      string    `json:"path,omitempty"`
	EnvVar    string    `json:"env_var,omitempty"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Logger writes structured audit events as JSON lines.
type Logger struct {
	w       io.Writer
	enabled bool
}

// New creates a Logger. If enabled is false all Log calls are no-ops.
func New(w io.Writer, enabled bool) *Logger {
	if w == nil {
		w = os.Stderr
	}
	return &Logger{w: w, enabled: enabled}
}

// Log writes an event to the underlying writer.
func (l *Logger) Log(e Event) error {
	if !l.enabled {
		return nil
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = l.w.Write(data)
	return err
}

// SecretFetched logs a successful secret retrieval.
func (l *Logger) SecretFetched(provider, path, envVar string) error {
	return l.Log(Event{
		Level:    LevelInfo,
		Action:   "secret_fetched",
		Provider: provider,
		Path:     path,
		EnvVar:   envVar,
	})
}

// SecretError logs a failed secret retrieval.
func (l *Logger) SecretError(provider, path, envVar string, fetchErr error) error {
	e := Event{
		Level:    LevelError,
		Action:   "secret_fetch_error",
		Provider: provider,
		Path:     path,
		EnvVar:   envVar,
	}
	if fetchErr != nil {
		e.Error = fetchErr.Error()
	}
	return l.Log(e)
}

// ProcessExec logs that a child process is about to be exec'd.
func (l *Logger) ProcessExec(argv []string) error {
	var cmd string
	if len(argv) > 0 {
		cmd = argv[0]
	}
	return l.Log(Event{
		Level:   LevelInfo,
		Action:  "process_exec",
		Message: cmd,
	})
}
