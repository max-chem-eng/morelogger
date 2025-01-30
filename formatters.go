package morelogger

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/max-chem-eng/morelogger/models"
)

type Formatter interface {
	Format(record models.LogRecord) ([]byte, error)
}

type JSONFormatter struct {
	bufPool sync.Pool
}

var byteBufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		bufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (jf *JSONFormatter) Format(record models.LogRecord) ([]byte, error) {
	buf := jf.bufPool.Get().(*bytes.Buffer)
	defer jf.bufPool.Put(buf)
	buf.Reset()

	buf.WriteByte('{')
	buf.WriteString(`"timestamp":"`)
	buf.WriteString(record.Timestamp.Format(TimeFormat))
	buf.WriteString(`",`)

	buf.WriteString(`"level":"`)
	buf.WriteString(record.Level.ToStr())
	buf.WriteString(`",`)

	buf.WriteString(`"message":`)
	encodeString(buf, record.Message)

	buf.WriteByte(',')
	for i, f := range record.Fields {
		buf.WriteByte('"')
		buf.WriteString(f.Key)
		buf.WriteString(`":`)
		encodeValue(buf, f.Value)
		if i < len(record.Fields)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')

	return buf.Bytes(), nil
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

func encodeString(w io.Writer, s string) {
	buf := byteBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write(strconv.AppendQuote(buf.Bytes(), s))
	w.Write(buf.Bytes())
	byteBufferPool.Put(buf)
}

// formatters.go
func encodeValue(w io.Writer, val interface{}) {
	switch v := val.(type) {
	case string:
		encodeString(w, v)
	case int:
		w.Write([]byte(strconv.Itoa(v)))
	case int32:
		w.Write([]byte(strconv.FormatInt(int64(v), 10)))
	case int64:
		w.Write([]byte(strconv.FormatInt(v, 10)))
	case float32:
		w.Write([]byte(strconv.FormatFloat(float64(v), 'f', -1, 32)))
	case float64:
		w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64)))
	case bool:
		if v {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	default:
		encodeString(w, fmt.Sprintf("%v", v))
	}
}
