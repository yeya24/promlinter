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

	issues := Run(fs, []*ast.File{file}, false)
	if len(issues) != 3 {
		t.Fatal()
	}

	if issues[0].Metric != "test_metric_name" && issues[0].Text != `counter metrics should have "_total" suffix` {
		t.Fatal()
	}

	if issues[1].Metric != "test_metric_total" && issues[0].Text != `no help text` {
		t.Fatal()
	}

	if issues[2].Metric != "foo" && issues[0].Text != `counter metrics should have "_total" suffix` {
		t.Fatal()
	}
}
