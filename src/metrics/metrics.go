package metrics

import "context"

type Metrics interface {
	Float64(ctx context.Context, metricName string, metric float64) error
	Int(ctx context.Context, metricName string, metric int) error
}
