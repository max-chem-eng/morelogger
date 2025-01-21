package morelogger

import (
	"fmt"
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/max-chem-eng/morelogger/models"
)

type Formatter interface {
	Format(record models.LogRecord) ([]byte, error)
}

type JSONFormatter struct {
	Indent bool
	// escapeHTML           bool
	// overrideRepeatedKeys bool
}

// Format implements Formatter for JSON output.
func (jf *JSONFormatter) Format(record models.LogRecord) ([]byte, error) {
	data := map[string]interface{}{
		"level":   record.Level.ToStr(),
		"message": record.Message,
	}
	if ShowTimestamp {
		data["timestamp"] = record.Timestamp.Format(TimeFormat)
	}

	for k, v := range record.Fields {
		data[k] = v
	}

	// Use encoding/json or any other library.
	if jf.Indent {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

type TextFormatter struct {
	ShowColor bool
}

// Format for a console-friendly output
func (cf *TextFormatter) Format(record models.LogRecord) ([]byte, error) {
	var sb strings.Builder

	if ShowTimestamp {
		sb.WriteString(record.Timestamp.Format(TimeFormat))
		sb.WriteString(" ")
	}

	if cf.ShowColor {
		sb.WriteString(colorForLevel(record.Level))
	}

	sb.WriteString(record.Level.ToStr())
	if cf.ShowColor {
		sb.WriteString(resetColor())
	}
	sb.WriteString(" ")

	sb.WriteString(record.Message)

	for k, v := range record.Fields {
		sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
	}

	return []byte(sb.String()), nil
}

func colorForLevel(level models.Level) string {
	switch level {
	case models.LevelDebug:
		return "\033[34m" // Blue
	case models.LevelWarn:
		return "\033[33m" // Yellow
	case models.LevelError, models.LevelFatal, models.LevelPanic:
		return "\033[31m" // Red
	default:
		return "\033[32m" // Green for Info
	}
}

func resetColor() string {
	return "\033[0m"
}
