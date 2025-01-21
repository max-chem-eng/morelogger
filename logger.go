package morelogger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/max-chem-eng/morelogger/models"
	"github.com/max-chem-eng/morelogger/pool"
)

type LoggerConfig struct {
	Level           models.Level
	AsyncBufferSize int
	Sampler         Sampler
	Sinks           []Sink
	AutoRotate      bool // TODO: handle log rotation
	RotateSize      int64
}

type LoggerOption func(*LoggerConfig)

func WithLevel(level models.Level) LoggerOption {
	return func(c *LoggerConfig) {
		c.Level = level
	}
}

func WithAsyncBuffer(size int) LoggerOption {
	return func(cfg *LoggerConfig) {
		cfg.AsyncBufferSize = size
	}
}

func WithSampling(sampler Sampler) LoggerOption {
	return func(cfg *LoggerConfig) {
		cfg.Sampler = sampler
	}
}

func WithSink(output io.Writer, formatter Formatter) LoggerOption {
	if formatter == nil {
		formatter = &TextFormatter{}
	}
	ws := NewWriterSink(output, formatter)
	return func(cfg *LoggerConfig) {
		cfg.Sinks = append(cfg.Sinks, ws)
	}
}

func WithFileSink(path string, formatter Formatter) LoggerOption {
	if formatter == nil {
		formatter = &TextFormatter{}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return func(cfg *LoggerConfig) {}
	}

	ws := NewWriterSink(file, formatter)
	return func(cfg *LoggerConfig) {
		cfg.Sinks = append(cfg.Sinks, ws)
	}
}

func (l *loggerImpl) WithCtx(ctx context.Context, fields ...Field) {
	ctx = l.with(ctx, fields...)
	l.ctx = ctx
}

func (l *loggerImpl) RemoveFromCtx(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}

	cData := getCtxData(ctx)
	for _, k := range keys {
		delete(cData.fields, k)
	}

	l.ctx = context.WithValue(ctx, ctxKey{}, cData)
}

type Field struct {
	Key   string
	Value interface{}
}

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

type ctxKey struct{}

type ctxData struct {
	fields map[string]interface{}
}

func getCtxData(ctx context.Context) *ctxData {
	data, ok := ctx.Value(ctxKey{}).(*ctxData)
	if !ok || data == nil {
		return &ctxData{fields: make(map[string]interface{})}
	}
	return data
}

// the sink interface is used to write log records to a destination
// such as stdout, stderr, a file, or a network connection.
type Sink interface {
	Write(record models.LogRecord) error
}

type WriterSink struct {
	output    io.Writer
	formatter Formatter
}

func NewWriterSink(output io.Writer, formatter Formatter) *WriterSink {
	if formatter == nil {
		formatter = defaultFormatter()
	}
	return &WriterSink{
		output:    output,
		formatter: formatter,
	}
}

func defaultFormatter() Formatter {
	return &TextFormatter{}
}

func (ws *WriterSink) Write(record models.LogRecord) error {
	data, err := ws.formatter.Format(record)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = ws.output.Write(data)
	return err
}

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Panic(msg string, fields ...Field)

	WithCtx(ctx context.Context, fields ...Field)
	RemoveFromCtx(ctx context.Context, keys ...string)
	Close()
}

type loggerImpl struct {
	level     models.Level
	sinks     []Sink
	mu        sync.Mutex
	asyncChan chan *models.LogRecord
	wg        sync.WaitGroup
	async     bool
	sampler   Sampler
	ctx       context.Context
}

func (l *loggerImpl) Debug(msg string, fields ...Field) {
	l.log(models.LevelDebug, msg, fields...)
}

func (l *loggerImpl) Info(msg string, fields ...Field) {
	l.log(models.LevelInfo, msg, fields...)
}

func (l *loggerImpl) Warn(msg string, fields ...Field) {
	l.log(models.LevelWarn, msg, fields...)
}

func (l *loggerImpl) Error(msg string, fields ...Field) {
	l.log(models.LevelError, msg, fields...)
}

func (l *loggerImpl) Fatal(msg string, fields ...Field) {
	l.log(models.LevelFatal, msg, fields...)
	os.Exit(1)
}

func (l *loggerImpl) Panic(msg string, fields ...Field) {
	l.log(models.LevelPanic, msg, fields...)
	panic(msg)
}

func (l *loggerImpl) log(level models.Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	// for now, record is not used so passing nil
	// this way we avoid getting the record from the pool
	// if a sampled that needs a record is implemented,
	// this will need to be updated, probably using a sampler.NeedRecord() method
	// to minimize allocations
	if l.sampler != nil && !l.sampler.Allow(nil) {
		return
	}

	// Acquire from pool
	record := pool.AcquireLogRecord()

	record.Timestamp = time.Now()
	record.Level = level
	record.Message = msg

	var ctx context.Context
	if l.ctx != nil {
		ctx = l.ctx
	} else {
		ctx = context.Background()
	}

	mergedFields := mergeFields(ctx, fields)
	for k, v := range mergedFields {
		record.Fields[k] = v
	}

	// If async is enabled, enqueue
	if l.async {
		select {
		case l.asyncChan <- record:
		default:
			// queue is full, drop the record
			// dropping is not ideal, but blocking could lead to deadlocks
		}
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// If not async, do sync logic
	// TODO: handle error
	for _, sink := range l.sinks {
		_ = sink.Write(*record)
		// if err != nil {
		// 	handleSynWriteError(sink, record)
		// }
	}
	pool.ReleaseLogRecord(record)
}

func (l *loggerImpl) with(ctx context.Context, fields ...Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	existingData := getCtxData(ctx)
	newData := &ctxData{
		fields: make(map[string]interface{}, len(existingData.fields)+len(fields)),
	}

	// Copy existing fields
	for k, v := range existingData.fields {
		newData.fields[k] = v
	}

	// Add new fields
	for _, f := range fields {
		newData.fields[f.Key] = f.Value
	}

	// Return a new context with updated data
	return context.WithValue(ctx, ctxKey{}, newData)
}

func mergeFields(ctx context.Context, extra []Field) map[string]interface{} {
	cData := getCtxData(ctx)
	merged := make(map[string]interface{}, len(cData.fields)+len(extra))
	for k, v := range cData.fields {
		merged[k] = v
	}

	for _, f := range extra {
		merged[f.Key] = f.Value
	}
	return merged
}

func New(options ...LoggerOption) Logger {
	cfg := &LoggerConfig{
		Level: models.LevelInfo,
	}
	for _, opt := range options {
		opt(cfg)
	}

	// If no sinks provided, log to console
	if len(cfg.Sinks) == 0 {
		cfg.Sinks = append(cfg.Sinks, NewWriterSink(os.Stdout, &TextFormatter{}))
	}

	l := &loggerImpl{
		level:   cfg.Level,
		sinks:   cfg.Sinks,
		sampler: cfg.Sampler,
	}

	// If async, start a goroutine to handle log records
	if cfg.AsyncBufferSize > 0 {
		l.async = true
		l.asyncChan = make(chan *models.LogRecord, cfg.AsyncBufferSize)
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			for record := range l.asyncChan {
				// Write record to each sink
				for _, sink := range l.sinks {
					_ = sink.Write(*record) // TODO: handle error
					// if err != nil {
					// 	handleAsynWriteError(sink, record)
					// }
				}
				pool.ReleaseLogRecord(record)
			}
		}()
	}

	return l
}

func (l *loggerImpl) Close() {
	close(l.asyncChan)
	l.wg.Wait()
}

func (l *loggerImpl) SetLevel(level models.Level) {
	l.level = level
}

// func handleAsynWriteError(sink Sink, record *models.LogRecord) {
// 	err := sink.Write(*record) // write synchronously
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "error writing to sink: %v\n", err)
// 	}
// }

// func handleSynWriteError(sink Sink, record *models.LogRecord) {
// 	err := sink.Write(*record)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "error writing to sink: %v\n", err)
//  }
// }
