package logger

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLoggerContext(ctx context.Context, debug bool) (context.Context, error) {
	zc := zap.NewProductionConfig()
	if debug {
		zc.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	zapLog, err := zc.Build()
	if err != nil {
		return nil, err
	}
	zapLog = zapLog.WithOptions()
	log := zapr.NewLogger(zapLog)
	loggerCtx := logr.NewContext(ctx, log)
	return loggerCtx, nil
}
