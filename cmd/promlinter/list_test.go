package main

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeya24/promlinter"
)

func TestLabel(t *testing.T) {
	fs := token.NewFileSet()

	metrics := promlinter.RunList(fs, findFiles([]string{"../../testdata/"}, fs), true)

	if len(metrics) != 10 {
		t.Fatal()
	}

	assert.Equal(t, []string{"namespace", "name"}, metrics[7].Labels())
	assert.Equal(t, []string{"namespace", "name", "const-label1=value1", "const-label2=?"}, metrics[8].Labels())
	assert.Equal(t, []string{"namespace", "name"}, metrics[9].Labels())
}
