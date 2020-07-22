package promlinter

const src = `
package promlinter

var testMetric = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "aaa",
	Help: "help",
})

var testMetric2 = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: test_total,
	Help: "",
}, []string{})
`

const bbb = `
package test

var testMetric2 = With(reg).NewHistogramVec(
		HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
		},
		[]string{"code", "handler", "method"},
	)
`
