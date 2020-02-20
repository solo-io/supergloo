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
	sink   *memorySink
	logger *zap.SugaredLogger
}

func (t *TestLogger) Logger() *zap.SugaredLogger {
	return t.logger
}

func (t *TestLogger) Sink() *memorySink {
	return t.sink
}

func (t *TestLogger) EXPECT() *logMatcher {
	return newLogMatcher(t.sink)
}

func NewTestLogger() *TestLogger {
	sink := &memorySink{new(zaptest.Buffer)}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), sink, zapcore.DebugLevel)
	logger := zap.New(core)
	return &TestLogger{
		sink:   sink,
		logger: logger.Sugar(),
	}
}

type logEntry map[string]interface{}

func newLogMatcher(sink *memorySink) *logMatcher {
	var entries []logEntry
	for _, v := range sink.Lines() {
		obj := logEntry{}
		Expect(json.Unmarshal([]byte(v), &obj)).NotTo(HaveOccurred())
		entries = append(entries, obj)
	}
	return &logMatcher{entries: entries}
}

type logMatcher struct {
	entries []logEntry
}

func (l *logMatcher) NumEntries(num int) *logMatcher {
	Expect(l.entries).To(HaveLen(num))
	return l
}

func (l *logMatcher) Entry(index int) *entryMatcher {
	Expect(l.entries).To(BeNumerically("<=", index), "index must be within range of the # of "+
		"avaiable entries")
	return &entryMatcher{entry: l.entries[index]}
}

func (l *logMatcher) LastEntry() *entryMatcher {
	return &entryMatcher{entry: l.entries[len(l.entries)-1]}
}

func (l *logMatcher) FirstEntry() *entryMatcher {
	return &entryMatcher{entry: l.entries[0]}
}

type entryMatcher struct {
	entry logEntry
}

func (e *entryMatcher) HaveMessage(msg string) *entryMatcher {
	Expect(e.entry).To(HaveKey(MsgEntryType))
	result := e.entry[MsgEntryType]
	Expect(result).To(Equal(msg))
	return e
}

func (e *entryMatcher) HaveError(err error) *entryMatcher {
	Expect(e.entry).To(HaveKey(ErrorEntryType))
	result := e.entry[ErrorEntryType]
	Expect(result).To(Equal(err.Error()))
	return e
}

func (e *entryMatcher) Have(key string, val interface{}) *entryMatcher {
	Expect(e.entry).To(HaveKeyWithValue(key, val))
	return e
}

func (e *entryMatcher) Level(level zapcore.Level) *entryMatcher {
	Expect(e.entry).To(HaveKey(LevelEntryType))
	result := e.entry[LevelEntryType]
	Expect(result).To(Equal(level.String()))
	return e
}

// for testing
type memorySink struct {
	*zaptest.Buffer
}

func (s *memorySink) Close() error { return nil }
