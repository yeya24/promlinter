package promlinter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "./testdata/testdata.go", nil, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	issues := RunLint(fs, []*ast.File{file}, Setting{Strict: false, DisabledLintFuncs: nil})

	if len(issues) != 7 {
		t.Fatalf("expect 7 issue, got %d, issues: %+#v", len(issues), issues)
	}

	for idx, iss := range issues {
		t.Logf("%d: %q: %s", idx, iss.Metric, iss.Pos)

		switch iss.Metric {
		case "kube_daemonset_labels", "test_metric_name", "foo":
			assert.Equal(t, iss.Text, `counter metrics should have "_total" suffix`)

		case "test_metric_total":
			assert.Equal(t, iss.Text, `no help text`)

		case "foo_bar_total":
			assert.Equal(t, iss.Text, `non-counter metrics should not have "_total" suffix`)

		case "kube_test_metric_count":
			assert.Equal(t, iss.Text, `non-histogram and non-summary metrics should not have "_count" suffix`)
		case "test_histogram_duration_seconds":
			assert.Equal(t, iss.Text, `metric name should not include type 'histogram'`)

		default:
			assert.Truef(t, false, "unexpected issue: %q", iss)
		}
	}
}
