package morelogger

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	defer stringBuilderPool.Put(sb)

	sb.WriteString("{")

	// Timestamp
	sb.WriteString(`"timestamp":"`)
	sb.WriteString(record.Timestamp.Format(time.RFC3339))
	sb.WriteString(`",`)

	// Level
	sb.WriteString(`"level":"`)
	sb.WriteString(record.Level.ToStr())
	sb.WriteString(`",`)

	// Message
	sb.WriteString(`"message":`)
	encodeString(sb, record.Message)
	sb.WriteString(`,`)

	// Fields
	// sb.WriteString(`"fields":{`)
	for i, f := range record.Fields {
		sb.WriteString(`"`)
		sb.WriteString(f.Key)
		sb.WriteString(`":`)
		encodeValue(sb, f.Value)
		if i < len(record.Fields)-1 {
			sb.WriteString(",")
		}
	}
	// sb.WriteString("}}")

	return []byte(sb.String()), nil
}

type TextFormatter struct {
	ShowColor bool
}

var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// Format for a console-friendly output
func (cf *TextFormatter) Format(record models.LogRecord) ([]byte, error) {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()

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

	for _, field := range record.Fields {
		sb.WriteString(fmt.Sprintf(" %s=%v", field.Key, field.Value))
	}

	data := []byte(sb.String())
	stringBuilderPool.Put(sb)
	return data, nil
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

func encodeString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\', '"':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		default:
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('"')
}

func encodeValue(sb *strings.Builder, v interface{}) {
	switch val := v.(type) {
	case string:
		encodeString(sb, val)
	case int, int32, int64, float32, float64, bool:
		fmt.Fprint(sb, val)
	default:
		// fallback if needed
		encodeString(sb, fmt.Sprintf("%v", val))
	}
}
