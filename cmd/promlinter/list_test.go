package main

import (
	"go/token"
	"testing"

	"github.com/yeya24/promlinter"
)

func TestLabel(t *testing.T) {
	fs := token.NewFileSet()

	metrics := promlinter.RunList(fs, findFiles([]string{"../../testdata/"}, fs), true)
	for _, m := range metrics {
		t.Logf("metric labels: %v", m.Labels())
	}

	if len(metrics) != 10 {
		t.Fatal()
	}
}
