package metrics

import (
	"context"
	"time"
)

type Metrics interface {
	Int(ctx context.Context, metricName string, metric int) error
	Duration(ctx context.Context, metricName string, metric time.Duration) error
	Success(ctx context.Context, metricName string, metric bool) error
}
