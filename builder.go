package morelogger

import (
	"io"
	"os"

	"github.com/max-chem-eng/morelogger/models"
)

type LoggerBuilder struct {
	config LoggerConfig
}

func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		config: LoggerConfig{
			Level: models.LevelInfo,
		},
	}
}

func (b *LoggerBuilder) WithLevel(level models.Level) *LoggerBuilder {
	b.config.Level = level
	return b
}

func (b *LoggerBuilder) WithSink(output io.Writer, formatter Formatter) *LoggerBuilder {
	if formatter == nil {
		formatter = &JSONFormatter{}
	}
	b.config.Sinks = append(b.config.Sinks, newWriterSink(output, formatter))
	return b
}

func (b *LoggerBuilder) ToFile(path string) *LoggerBuilder {
	file, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	b.config.Sinks = append(b.config.Sinks, newWriterSink(file, NewJSONFormatter()))
	return b
}

func (b *LoggerBuilder) WithAsyncBuffer(size int) *LoggerBuilder {
	b.config.AsyncBufferSize = size
	return b
}

func (b *LoggerBuilder) Build() Logger {
	return New(b.config)
}
