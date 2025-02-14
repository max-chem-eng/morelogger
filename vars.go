package morelogger

import "github.com/max-chem-eng/morelogger/models"

var (
	TimeFormat    = "2006-01-02 15:04:05.000"
	ShowTimestamp = true
	// LevelErrorTag = "ERROR"
	TimestampKey = "timestamp"
	LevelKey     = "level"
	MessageKey   = "message"
)

// defaults

func defaultFormatter() Formatter {
	return &JSONFormatter{}
}

func defaultLevel() models.Level {
	return models.LevelInfo
}
