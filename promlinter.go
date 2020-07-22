package promlinter

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
	dto "github.com/prometheus/client_model/go"
)

var metricsType map[string]dto.MetricType

func init() {
	metricsType = map[string]dto.MetricType{
		"NewCounter":      dto.MetricType_COUNTER,
		"NewCounterVec":   dto.MetricType_COUNTER,
		"NewGauge":        dto.MetricType_GAUGE,
		"NewGaugeVec":     dto.MetricType_GAUGE,
		"NewHistogram":    dto.MetricType_HISTOGRAM,
		"NewHistogramVec": dto.MetricType_HISTOGRAM,
		"NewSummary":      dto.MetricType_SUMMARY,
		"NewSummaryVec":   dto.MetricType_SUMMARY,
	}
}

// Issue contains a description of linting error
type Issue struct {
	Pos    token.Position
	Metric string
	Text   string
}

// Run runs this linter on the provided code.
func Run(file *ast.File, fset *token.FileSet) []Issue {
	issues := []Issue{}

	var (
		name         string
		promautoName string
	)
	for _, pkg := range file.Imports {
		switch pkg.Path.Value {
		case `"github.com/prometheus/client_golang/prometheus"`:
			if pkg.Name != nil {
				name = pkg.Name.Name
			} else {
				name = "prometheus"
			}

		case `"github.com/prometheus/client_golang/prometheus/promauto"`:
			if pkg.Name != nil {
				promautoName = pkg.Name.Name
			} else {
				promautoName = "promauto"
			}
		}
	}

	l := &metric{
		fileName:     file.Name.Name,
		packageName:  name,
		promautoName: promautoName,
	}

	ast.Walk(l, file)

	if len(l.Metrics) > 0 {
		lint := promlint.NewWithMetricFamilies(l.Metrics)
		problems, err := lint.Lint()
		if err != nil {
			panic(err)
		}

		for _, p := range problems {
			issues = append(issues, Issue{
				Pos:    token.Position{},
				Metric: p.Metric,
				Text:   p.Text,
			})
		}
	}
	return issues
}

type metric struct {
	fileName     string
	packageName  string
	promautoName string
	Metrics      []*dto.MetricFamily
}

func (l *metric) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return l
	}
	ce, ok := n.(*ast.CallExpr)
	if !ok {
		return l
	}
	se, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return l
	}

	switch se.X.(type) {
	case *ast.Ident:
		id := se.X.(*ast.Ident)
		if id.Name != l.packageName && id.Name != l.promautoName {
			return l
		}

	case *ast.CallExpr:
		innerCE := se.X.(*ast.CallExpr)

		switch innerCE.Fun.(type) {
		case *ast.Ident:
			innerID := innerCE.Fun.(*ast.Ident)
			if l.promautoName != "." || innerID.Name != "With" {
				return l
			}

		case *ast.SelectorExpr:
			funSE := innerCE.Fun.(*ast.SelectorExpr)
			if funSE.Sel.Name != "With" {
				return l
			}

			funXID, ok := funSE.X.(*ast.Ident)
			if !ok {
				return l
			}

			if funXID.Name != l.promautoName {
				return l
			}
		}
	}

	metricType, ok := metricsType[se.Sel.Name]
	if !ok {
		return l
	}

	// Check first arg, that should have basic lit with capital
	if len(ce.Args) < 1 {
		return l
	}
	opts, help := parseOpts(ce.Args[0])
	if opts == nil {
		return l
	}

	currentMetric := dto.MetricFamily{
		Type: &metricType,
		Help: help,
	}

	metricName := prometheus.BuildFQName(opts.namespace, opts.sub, opts.name)
	currentMetric.Name = &metricName

	l.Metrics = append(l.Metrics, &currentMetric)
	return l
}

func parseOpts(n ast.Node) (*opt, *string) {
	metricOption := &opt{}
	var help *string
	if option, ok := n.(*ast.CompositeLit); ok {
		for _, elt := range option.Elts {
			kvExpr, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			object, ok := kvExpr.Key.(*ast.Ident)
			if !ok {
				continue
			}

			stringLiteral, ok := parseValue(kvExpr.Value)
			if !ok {
				continue
			}

			switch object.Name {
			case "Namespace":
				metricOption.namespace = stringLiteral
			case "Subsystem":
				metricOption.sub = stringLiteral
			case "Name":
				metricOption.name = stringLiteral
			case "Help":
				help = &stringLiteral
			}
		}

		return metricOption, help
	}

	if identName, ok := n.(*ast.Ident); ok {
		fmt.Println(identName.Obj)
	}

	return nil, nil
}

func parseValue(n ast.Node) (string, bool) {
	switch t := n.(type) {
	case *ast.BasicLit:
		return mustUnquote(t.Value), true
	case *ast.Ident:
		if vs, ok := t.Obj.Decl.(*ast.ValueSpec); !ok {
			return "", false
		} else {
			return parseValue(vs)
		}
	case *ast.ValueSpec:
		if len(t.Values) == 0 {
			return "", false
		}
		return parseValue(t.Values[0])
	default:
		return "", false
	}
}

func mustUnquote(str string) string {
	stringLiteral, err := strconv.Unquote(str)
	if err != nil {
		panic(err)
	}

	return stringLiteral
}

type opt struct {
	namespace string
	sub       string
	name      string
	help      string
}
