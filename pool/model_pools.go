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
	for _, f := range r.Fields {
		ReleaseField(f)
	}
	r.Fields = r.Fields[:0]

	r.Message = ""
	r.Level = 0
	r.Timestamp = time.Time{}
	logRecordPool.Put(r)
}

var fieldPool = sync.Pool{
	New: func() interface{} {
		f := &models.Field{
			Key:   "",
			Value: nil,
		}
		return f
	},
}

func AcquireField() *models.Field {
	return fieldPool.Get().(*models.Field)
}

func ReleaseField(r *models.Field) {
	r.Key = ""
	r.Value = nil
	fieldPool.Put(r)
}

func AcquireFieldKV(key string, value interface{}) *models.Field {
	f := AcquireField()
	f.Key = key
	f.Value = value
	return f
}
