// examples for testing

package testdata

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/component-base/metrics"
	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	descDaemonSetLabelsName = "kube_daemonset_labels"
	descDaemonSetLabelsHelp = "Kubernetes labels converted to Prometheus labels."
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

	// NewCounterFunc, should have _total suffix
	_ = promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "foo",
		Help: "bar",
	}, func() float64 {
		return 1
	})

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

	// good: const labels
	var (
		constLabelVal2 = "value2"
	)
	descConstLabel := prometheus.NewDesc(
		"prometheus_operator_spec_replicas",
		"Number of expected replicas for the object.",
		[]string{
			"namespace",
			"name",
		},
		map[string]string{
			"const-label1": "value1",
			"const-label2": constLabelVal2,
		},
	)
	ch <- prometheus.MustNewConstMetric(descConstLabel, prometheus.GaugeValue, 1)

	// support using BuildFQName to generate fqName here.
	// bad metric, gauge shouldn't have _total
	ch <- prometheus.MustNewConstMetric(prometheus.NewDesc(
		prometheus.BuildFQName("foo", "bar", "total"),
		"Number of expected replicas for the object.",
		[]string{
			"namespace",
			"name",
		}, nil), prometheus.GaugeValue, 1)

	// support detecting kubernetes metrics
	kubeMetricDesc := metrics.NewDesc(
		"kube_test_metric_count",
		"Gauge Help",
		[]string{}, nil, metrics.STABLE, "",
	)
	ch <- metrics.NewLazyConstMetric(kubeMetricDesc, metrics.GaugeValue, 1)

	// bad
	_ = metrics.NewHistogram(&metrics.HistogramOpts{
		Name: "test_histogram_duration_seconds",
		Help: "",
	})

	// https://github.com/prometheus/mysqld_exporter/blob/master/collector/engine_innodb.go#L78-L82
	// This is not supported because we cannot infer what newDesc is doing before runtime.
	ch <- prometheus.MustNewConstMetric(
		newDesc("innodb", "queries_inside_innodb", "Queries inside InnoDB."),
		prometheus.GaugeValue,
		1,
	)

	// metrics for kube-state-metrics
	_ = []generator.FamilyGenerator{
		// good
		*generator.NewFamilyGenerator(
			"kube_daemonset_created",
			"foo",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			descDaemonSetLabelsName,
			descDaemonSetLabelsHelp,
			metric.Counter,
			"",
			nil,
		),
	}

	// We skip linting these case when metric name is not set.
	promauto.With(nil).NewCounter(prometheus.CounterOpts{})
	promauto.With(nil).NewCounterVec(prometheus.CounterOpts{}, nil)
}

func newDesc(subsystem, name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("foo", subsystem, name),
		help, nil, nil,
	)
}
