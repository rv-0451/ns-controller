package metrics

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

var metricslog = log.Log.WithName("[metrics]")

type MetricsProvider interface {
	QueryMemoryOverprovisioning() (float64, error)
}

func NewMetricsProvider() MetricsProvider {
	log.FromContext(context.Background())
	return NewKubeStateMetrics()
}

type MetricNotFoundError struct {
	Msg string
}

func (e *MetricNotFoundError) Error() string {
	return fmt.Sprintf("Metric not found: %s", e.Msg)
}
