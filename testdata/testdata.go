// examples for testing

package testdata

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	descDaemonSetLabelsName = "kube_daemonset_labels"
	descDaemonSetLabelsHelp = "Kubernetes labels converted to Prometheus labels."

	_ = []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kube_daemonset_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_current_number_scheduled",
			"The number of nodes running at least one daemon pod and are supposed to.",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_desired_number_scheduled",
			"The number of nodes that should be running the daemon pod.",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_number_available",
			"The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_number_misscheduled",
			"The number of nodes running a daemon pod but are not supposed to.",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_number_ready",
			"The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready.",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_number_unavailable",
			"The number of nodes that should be running the daemon pod and have none of the daemon pod running and available",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_observed_generation",
			"The most recent generation observed by the daemon set controller.",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_status_updated_number_scheduled",
			"The total number of nodes that are running updated daemon pod",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			"kube_daemonset_metadata_generation",
			"Sequence number representing a specific generation of the desired state.",
			metric.Gauge,
			"",
			nil,
		),
		*generator.NewFamilyGenerator(
			descDaemonSetLabelsName,
			descDaemonSetLabelsHelp,
			metric.Gauge,
			"",
			nil,
		),
	}
)

func main() {
	ch := make(chan<- prometheus.Metric)

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
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1)

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
}

func newDesc(subsystem, name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("foo", subsystem, name),
		help, nil, nil,
	)
}
