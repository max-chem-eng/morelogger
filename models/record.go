package models

import "time"

type LogRecord struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Fields    []Field
}
