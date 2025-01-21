package pool

import (
	"sync"
	"time"

	"github.com/max-chem-eng/morelogger/models"
)

// LogRecordPool wraps a sync.Pool for reusing LogRecord objects.
// This is used to reduce memory allocations and GC pressure.
// The pool is safe for concurrent use.
// The function in New creates a blank record with empty
// Fields map.
var LogRecordPool = sync.Pool{
	New: func() interface{} {
		return &models.LogRecord{
			Fields: make(map[string]interface{}),
		}
	},
}

// AcquireLogRecord fetches a *LogRecord from the pool.
// TODO: rename to GetLogRecord
func AcquireLogRecord() *models.LogRecord {
	return LogRecordPool.Get().(*models.LogRecord)
}

// ReleaseLogRecord resets the record fields and puts it back into the pool.
func ReleaseLogRecord(r *models.LogRecord) {
	// Clear out fields to avoid carrying over data.
	for k := range r.Fields {
		delete(r.Fields, k)
	}
	r.Message = ""
	r.Level = 0
	r.Timestamp = time.Time{}

	LogRecordPool.Put(r)
}
