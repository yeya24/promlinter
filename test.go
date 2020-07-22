package promlinter

import (
	aaa "github.com/prometheus/client_golang/prometheus"
)

var testMetric = aaa.NewCounter(aaa.CounterOpts{
	Name: "'aaa",
	Help: "bb",
})
