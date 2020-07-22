package promlinter

import (
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"runtime"
	"strconv"
	"sync"
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

// Settings contains linter settings.
type Settings struct {
}

// Issue contains a description of linting error and a possible replacement.
type Issue struct {
	Pos    token.Position
	Metric string
	Text   string
}

type visitor struct {
	pkg          *packages.Package
	Metrics      []*dto.MetricFamily
}

type Checker struct {
}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) load(paths ...string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedDeps | packages.NeedImports | packages.NeedSyntax | packages.NeedTypes,
		Tests:      false,
		BuildFlags: nil,
	}
	return packages.Load(cfg, paths...)
}

func (c *Checker) CheckPackages(paths ...string) error {
	pkgs, err := c.load(paths...)
	if err != nil {
		return err
	}
	// Check for errors in the initial packages.
	work := make(chan *packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return fmt.Errorf("errors while loading package %s: %v", pkg.ID, pkg.Errors)
		}
		work <- pkg
	}
	close(work)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for pkg := range work {

				v := &visitor{
					pkg: pkg,
				}

				for _, astFile := range v.pkg.Syntax {
					s := v.pkg.Types.Scope()
					fmt.Println(s.Names())
					fmt.Println(s.Lookup("ccc").Type())
					ast.Walk(v, astFile)
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

//// Run runs this linter on the provided code.
//func Run(file *ast.File, fset *token.FileSet, settings Settings) []Issue {
//	issues := []Issue{}
//
//	var (
//		name         string
//		promautoName string
//	)
//	for _, pkg := range file.Imports {
//		switch pkg.Path.Value {
//		case `"github.com/prometheus/client_golang/prometheus"`:
//			if pkg.Name != nil {
//				name = pkg.Name.Name
//			} else {
//				name = "prometheus"
//			}
//
//		case `"github.com/prometheus/client_golang/prometheus/promauto"`:
//			if pkg.Name != nil {
//				promautoName = pkg.Name.Name
//			} else {
//				promautoName = "promauto"
//			}
//		}
//	}
//
//	l := &metric{
//		fileName:     file.Name.Name,
//		packageName:  name,
//		promautoName: promautoName,
//	}
//
//	ast.Walk(l, file)
//
//	if len(l.Metrics) > 0 {
//		lint := promlint.NewWithMetricFamilies(l.Metrics)
//		problems, err := lint.Lint()
//		if err != nil {
//			panic(err)
//		}
//
//		for _, p := range problems {
//			issues = append(issues, Issue{
//				Pos:    token.Position{},
//				Metric: p.Metric,
//				Text:   p.Text,
//			})
//		}
//	}
//	return issues
//}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return v
	}

	//switch stmt := n.(type) {
	//case *ast.CallExpr:
	//	//if t, ok := metricsType[stmt.Fun]
	//	fmt.Println(stmt)
	//
	//case *ast.SelectorExpr:
	//	fmt.Println(stmt)
	//
	//default:
	//	fmt.Println("%T", n)
	//}

	call, ok := n.(*ast.CallExpr)
	if !ok {
		return v
	}

	if len(call.Args) < 1 || len(call.Args) > 2 {
		return v
	}

	switch stmt := call.Fun.(type) {
	case *ast.CallExpr:
		//if t, ok := metricsType[stmt.Fun]
		fmt.Println(stmt)

	case *ast.Ident:
		if t, ok := metricsType[stmt.Name]; !ok {
			return v
		} else {
			fmt.Println(t)
		}
		//fmt.Println(call.Args[0])
		v.parseOpts(call.Args[0])

	case *ast.SelectorExpr:
		if t, ok := metricsType[stmt.Sel.Name]; !ok {
			return v
		} else {
			fmt.Println(t)
		}
		fmt.Println(stmt)

	default:
	}

	return v
}

func (v *visitor) parseOpts(expr ast.Node) {
	metricOption := new(opt)
	switch stmt := expr.(type) {
	case *ast.Ident:
	case *ast.CompositeLit:
		for _, elt := range stmt.Elts {
			kvExpr, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			object, ok := kvExpr.Key.(*ast.Ident)
			if !ok {
				continue
			}

			stringLiteral, ok := v.parseValue(kvExpr.Value)
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
				//help = &stringLiteral
			}
		}

	default:
		fmt.Println(stmt)
	}
	//switch t := v.pkg.TypesInfo.Types[call].Type.(type) {
	//default:
	//	fmt.Println(v.pkg.TypesInfo.Types[call].Value)
	//	fmt.Println(t)
	//}
}

//func (l *visitor) Visit(n ast.Node) ast.Visitor {
//	if n == nil {
//		return l
//	}
//
//	ce, ok := n.(*ast.CallExpr)
//	if !ok {
//		return l
//	}
//	se, ok := ce.Fun.(*ast.SelectorExpr)
//	if !ok {
//		return l
//	}
//
//	switch t := se.X.(type) {
//	case *ast.Ident:
//		aaa, ok := l.pkg.TypesInfo.Types[t]
//		if ok {
//			fmt.Println(aaa)
//		}
//		//id := se.X.(*ast.Ident)
//		//if id.Name != l.packageName && id.Name != l.promautoName {
//		//	return l
//		//}
//
//	case *ast.CallExpr:
//		innerCE := se.X.(*ast.CallExpr)
//
//		switch innerCE.Fun.(type) {
//		case *ast.Ident:
//			//innerID := innerCE.Fun.(*ast.Ident)
//			//if l.promautoName != "." || innerID.Name != "With" {
//			//	return l
//			//}
//
//		case *ast.SelectorExpr:
//			funSE := innerCE.Fun.(*ast.SelectorExpr)
//			if funSE.Sel.Name != "With" {
//				return l
//			}
//
//			//funXID, ok := funSE.X.(*ast.Ident)
//			//if !ok {
//			//	return l
//			//}
//			//
//			//if funXID.Name != l.promautoName {
//			//	return l
//			//}
//		}
//	}
//
//	metricType, ok := metricsType[se.Sel.Name]
//	if !ok {
//		return l
//	}
//
//	// Check first arg, that should have basic lit with capital
//	if len(ce.Args) < 1 {
//		return l
//	}
//	opts, help := parseOpts(l.pkg.Fset, ce.Args[0])
//	if opts == nil {
//		return l
//	}
//
//	currentMetric := dto.MetricFamily{
//		Type: &metricType,
//		Help: help,
//	}
//	//// parse Namespace Subsystem Name Help
//	//var namespace, subsystem, name string
//	//for _, elt := range opts.Elts {
//	//	expr, ok := elt.(*ast.KeyValueExpr)
//	//	if !ok {
//	//		continue
//	//	}
//	//	object, ok := expr.Key.(*ast.Ident)
//	//	if !ok {
//	//		continue
//	//	}
//	//	value, ok := expr.Value.(*ast.BasicLit)
//	//	if !ok {
//	//		continue
//	//	}
//	//
//	//	stringLiteral := mustUnquote(value.Value)
//	//	switch object.Name {
//	//	case "Namespace":
//	//		namespace = stringLiteral
//	//	case "Subsystem":
//	//		subsystem = stringLiteral
//	//	case "Name":
//	//		name = stringLiteral
//	//	case "Help":
//	//		currentMetric.Help = &stringLiteral
//	//	}
//	//}
//
//	metricName := prometheus.BuildFQName(opts.namespace, opts.sub, opts.name)
//	currentMetric.Name = &metricName
//
//	l.Metrics = append(l.Metrics, &currentMetric)
//	return l
//}
//
//func parseOpts(fs *token.FileSet, n ast.Node) (*opt, *string) {
//	metricOption := &opt{}
//	var help *string
//	if option, ok := n.(*ast.CompositeLit); ok {
//		for _, elt := range option.Elts {
//			kvExpr, ok := elt.(*ast.KeyValueExpr)
//			if !ok {
//				continue
//			}
//			object, ok := kvExpr.Key.(*ast.Ident)
//			if !ok {
//				continue
//			}
//
//			stringLiteral, ok := parseValue(kvExpr.Value)
//			if !ok {
//				continue
//			}
//
//			switch object.Name {
//			case "Namespace":
//				metricOption.namespace = stringLiteral
//			case "Subsystem":
//				metricOption.sub = stringLiteral
//			case "Name":
//				metricOption.name = stringLiteral
//			case "Help":
//				help = &stringLiteral
//			}
//		}
//
//		return metricOption, help
//	}
//
//	if identName, ok := n.(*ast.Ident); ok {
//		fmt.Println(identName.Obj)
//	}
//
//	return nil, nil
//}

func parseIdentifier() {

}

func (v *visitor) parseValue(n ast.Node) (string, bool) {
	switch t := n.(type) {
	case *ast.BasicLit:
		return mustUnquote(t.Value), true
	case *ast.Ident:
		//if vs, ok := t.Obj.Decl.(*ast.ValueSpec); !ok {
		//	return "", false
		//} else {
		//	return parseValue(vs)
		//}
		fmt.Println(v.pkg.Imports)
	case *ast.ValueSpec:
		if len(t.Values) == 0 {
			return "", false
		}
		return v.parseValue(t.Values[0])
	case *ast.SelectorExpr:
		return v.parseValue(t.Sel)
	default:
		return "", false
	}

	return "", false
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
