package metrics

import "context"

type Metrics interface {
	Int(ctx context.Context, metricName string, metric int) error
}
