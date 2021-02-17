package promlinter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestRun(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "./testdata/testdata.go", nil, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	issues := RunLint(fs, []*ast.File{file}, Setting{Strict: false, DisabledLintFuncs: nil})
	if len(issues) != 7 {
		t.Fatal()
	}

	if issues[0].Metric != "kube_daemonset_labels" && issues[0].Text != `counter metrics should have "_total" suffix` {
		t.Fatal()
	}

	if issues[1].Metric != "test_metric_name" && issues[1].Text != `counter metrics should have "_total" suffix` {
		t.Fatal()
	}

	if issues[2].Metric != "test_metric_total" && issues[2].Text != `no help text` {
		t.Fatal()
	}

	if issues[3].Metric != "foo" && issues[3].Text != `counter metrics should have "_total" suffix` {
		t.Fatal()
	}

	if issues[4].Metric != "foo_bar_total" && issues[4].Text != `non-counter metrics should not have "_total" suffix` {
		t.Fatal()
	}

	if issues[5].Metric != "kube_test_metric_count" && issues[5].Text != `non-histogram and non-summary metrics should not have "_count" suffix` {
		t.Fatal()
	}

	if issues[6].Metric != "test_histogram_duration_seconds" && issues[6].Text != `metric name should not include type 'histogram'` {
		t.Fatal()
	}
}
