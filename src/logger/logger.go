package logger

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func NewLoggerContext(ctx context.Context) (context.Context, error) {
	zapLog, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	log := zapr.NewLogger(zapLog)
	loggerCtx := logr.NewContext(ctx, log)
	return loggerCtx, nil
}
