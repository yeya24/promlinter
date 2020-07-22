// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package thanos

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yeya24/promlinter/test"
)

const (
	aaa = "this is a test"
	bbb = "bbb"
)

var (
	ccc = prometheus.HistogramOpts{
		Name: ddd,
		Help: bbb,
	}
	ddd = aaa
)

func main() {
	//a := prometheus.HistogramOpts{
	//	Name: aaa,
	//}
	_ = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: test.G,
			Help: bbb,
		},
		[]string{},
	)
}
