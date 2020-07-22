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

	// good
	f := promauto.With(prometheus.NewRegistry())
	_ = f.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_metric_total",
			Help: "",
		},
		[]string{},
	)

	// good
	_ = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_metric_total",
			Help: "",
		},
		[]string{},
	)

	// good
	desc := prometheus.NewDesc(
		"prometheus_operator_spec_replicas",
		"Number of expected replicas for the object.",
		[]string{
			"namespace",
			"name",
		}, nil,
	)
	ch := make(chan<- prometheus.Metric)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1)
}
