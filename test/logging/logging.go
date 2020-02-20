package test_logging

import (
	"encoding/json"

	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

const (
	MsgEntryType   = "msg"
	LevelEntryType = "level"
	ErrorEntryType = "error"
)

type TestLogger struct {
	sink   *MemorySink
	logger *zap.SugaredLogger
}

func (t *TestLogger) Logger() *zap.SugaredLogger {
	return t.logger
}

func (t *TestLogger) Sink() *MemorySink {
	return t.sink
}

func (t *TestLogger) EXPECT() *LogMatcher {
	return newLogMatcher(t.sink)
}

func NewTestLogger() *TestLogger {
	sink := &MemorySink{new(zaptest.Buffer)}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), sink, zapcore.DebugLevel)
	logger := zap.New(core)
	return &TestLogger{
		sink:   sink,
		logger: logger.Sugar(),
	}
}

type logEntry map[string]interface{}

func newLogMatcher(sink *MemorySink) *LogMatcher {
	var entries []logEntry
	for _, v := range sink.Lines() {
		obj := logEntry{}
		Expect(json.Unmarshal([]byte(v), &obj)).NotTo(HaveOccurred())
		entries = append(entries, obj)
	}
	return &LogMatcher{entries: entries}
}

type LogMatcher struct {
	entries []logEntry
}

func (l *LogMatcher) NumEntries(num int) *LogMatcher {
	Expect(l.entries).To(HaveLen(num))
	return l
}

func (l *LogMatcher) Entry(index int) *EntryMatcher {
	Expect(l.entries).To(BeNumerically("<=", index), "index must be within range of the # of "+
		"avaiable entries")
	return &EntryMatcher{entry: l.entries[index]}
}

func (l *LogMatcher) LastEntry() *EntryMatcher {
	return &EntryMatcher{entry: l.entries[len(l.entries)-1]}
}

func (l *LogMatcher) FirstEntry() *EntryMatcher {
	return &EntryMatcher{entry: l.entries[0]}
}

type EntryMatcher struct {
	entry logEntry
}

func (e *EntryMatcher) HaveMessage(msg string) *EntryMatcher {
	Expect(e.entry).To(HaveKey(MsgEntryType))
	result := e.entry[MsgEntryType]
	Expect(result).To(Equal(msg))
	return e
}

func (e *EntryMatcher) HaveError(err error) *EntryMatcher {
	Expect(e.entry).To(HaveKey(ErrorEntryType))
	result := e.entry[ErrorEntryType]
	Expect(result).To(Equal(err.Error()))
	return e
}

func (e *EntryMatcher) Have(key string, val interface{}) *EntryMatcher {
	Expect(e.entry).To(HaveKeyWithValue(key, val))
	return e
}

func (e *EntryMatcher) Level(level zapcore.Level) *EntryMatcher {
	Expect(e.entry).To(HaveKey(LevelEntryType))
	result := e.entry[LevelEntryType]
	Expect(result).To(Equal(level.String()))
	return e
}

// for testing
type MemorySink struct {
	*zaptest.Buffer
}

func (s *MemorySink) Close() error { return nil }
