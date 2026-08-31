package interfaces

import "github.com/divibisoul/Orquestrador-/compute/transcendental/core"

type MetricsProvider interface {
	GetLastMetrics() core.Metrics
	GetMetricsHistory(limit int) []core.Metrics
}
