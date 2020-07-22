// examples for testing

package testdata

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func main() {
	// counter metric should have _total suffix
	_ = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_metric_name",
			Help: "test help text",
		},
		[]string{},
	)

	// no help text
	_ = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_metric_total",
		},
		[]string{},
	)
}
