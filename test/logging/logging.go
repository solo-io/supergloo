package test_logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

// for testing
type MemorySink struct {
	*zaptest.Buffer
}

func (s *MemorySink) Close() error { return nil }

func NewInMemoryLogger() (*MemorySink, *zap.SugaredLogger) {
	memorySink := &MemorySink{new(zaptest.Buffer)}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), memorySink, zapcore.DebugLevel)
	logger := zap.New(core)
	return memorySink, logger.Sugar()
}
