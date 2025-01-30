package models

import "time"

type Level int8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

var levelsMap = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
	LevelPanic: "PANIC",
}

func (l *Level) ToStr() string {
	return levelsMap[*l]
}

type Field struct {
	Key   string
	Value interface{}
}

type LogRecord struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Fields    []*Field
}
