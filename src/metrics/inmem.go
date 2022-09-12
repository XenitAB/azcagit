package metrics

import (
	"context"
	"time"
)

type InMemMetrics struct {
	intMetrics      []int
	durationMetrics []time.Duration
	successMetrics  []bool
}

func NewInMemMetrics() *InMemMetrics {
	return &InMemMetrics{}
}

var _ Metrics = (*InMemMetrics)(nil)

func (m *InMemMetrics) Int(ctx context.Context, metricName string, metric int) error {
	m.intMetrics = append(m.intMetrics, metric)
	return nil
}

func (m *InMemMetrics) IntStats() []int {
	return m.intMetrics
}

func (m *InMemMetrics) Duration(ctx context.Context, metricName string, metric time.Duration) error {
	m.durationMetrics = append(m.durationMetrics, metric)
	return nil
}

func (m *InMemMetrics) DurationStats() []time.Duration {
	return m.durationMetrics
}

func (m *InMemMetrics) Success(ctx context.Context, metricName string, metric bool) error {
	m.successMetrics = append(m.successMetrics, metric)
	return nil
}

func (m *InMemMetrics) SuccessStats() []bool {
	return m.successMetrics
}

func (m *InMemMetrics) Reset() {
	m.intMetrics = []int{}
	m.durationMetrics = []time.Duration{}
	m.successMetrics = []bool{}
}
