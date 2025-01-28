package pool

import (
	"sync"
	"time"

	"github.com/max-chem-eng/morelogger/models"
)

var logRecordPool = sync.Pool{
	New: func() interface{} {
		r := &models.LogRecord{
			Fields: nil,
		}
		return r
	},
}

func AcquireLogRecord() *models.LogRecord {
	return logRecordPool.Get().(*models.LogRecord)
}

func ReleaseLogRecord(r *models.LogRecord) {
	r.Fields = nil
	r.Message = ""
	r.Level = 0
	r.Timestamp = time.Time{}
	logRecordPool.Put(r)
}
