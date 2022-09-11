package metrics

import "context"

type InMemMetrics struct {
	float64Metrics []float64
	intMetrics     []int
}

func NewInMemMetrics() *InMemMetrics {
	return &InMemMetrics{}
}

var _ Metrics = (*InMemMetrics)(nil)

func (m *InMemMetrics) Float64Metrics() []float64 {
	return m.float64Metrics
}

func (m *InMemMetrics) Float64(ctx context.Context, metricName string, metric float64) error {
	m.float64Metrics = append(m.float64Metrics, metric)
	return nil
}

func (m *InMemMetrics) IntStats() []int {
	return m.intMetrics
}

func (m *InMemMetrics) Int(ctx context.Context, metricName string, metric int) error {
	m.intMetrics = append(m.intMetrics, metric)
	return nil
}

func (m *InMemMetrics) Reset() {
	m.float64Metrics = []float64{}
	m.intMetrics = []int{}
}
