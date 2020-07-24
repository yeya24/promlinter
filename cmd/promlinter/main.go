package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/yeya24/promlinter"
)

const help = `Prometheus metrics linter for Go code.

This tool can cover most of the patterns of metrics naming issues, but it cannot detect metric values that can only be determined in the runtime.

By default it doesn't output parsing failures, if you want to see them, you can add --strict flag to enable it.

It is also supported to disable the lint functions using repeated flag --disable. Current supported functions are:

	[Help]: Help detects issues related to the help text for a metric.

	[MetricUnits]: MetricUnits detects issues with metric unit names.

	[Counter]: Counter detects issues specific to counters, as well as patterns that should only be used with counters.

	[HistogramSummaryReserved]: HistogramSummaryReserved detects when other types of metrics use names or labels reserved for use by histograms and/or summaries.

	[MetricTypeInName]: MetricTypeInName detects when metric types are included in the metric name.

	[ReservedChars]: ReservedChars detects colons in metric names.

	[CamelCase]: CamelCase detects metric names and label names written in camelCase.

	[UnitAbbreviations]: UnitAbbreviations detects abbreviated units in the metric name.
`

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), help)
	app.Version("v0.0.1")
	app.HelpFlag.Short('h')

	paths := app.Arg("files", "Files to lint.").Strings()
	strict := app.Flag("strict", "Strict mode. If true, linter will output more issues including parsing failures.").
		Default("false").Short('s').Bool()
	disableLintFuncs := app.Flag("disable", "Disable lint functions (repeated)."+
		"Supported options: Help, Counter, MetricUnits, HistogramSummaryReserved, MetricTypeInName, "+
		"ReservedChars, CamelCase, UnitAbbreviations").Short('d').Enums(promlinter.LintFuncNames...)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var files []*ast.File
	fileSet := token.NewFileSet()

	for _, path := range *paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Exit(1)
		}
		for f := range findFiles(path) {
			file, err := parser.ParseFile(fileSet, f, nil, parser.AllErrors)
			if err != nil {
				os.Exit(1)
			}
			files = append(files, file)
		}
	}

	setting := promlinter.Setting{Strict: *strict, DisabledLintFuncs: *disableLintFuncs}
	for _, iss := range promlinter.Run(fileSet, files, setting) {
		fmt.Printf("%s %s %s\n", iss.Pos, iss.Metric, iss.Text)
	}
}

func findFiles(root string) chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			sep := string(filepath.Separator)
			if strings.HasPrefix(path, "vendor"+sep) || strings.Contains(path, sep+"vendor"+sep) {
				return nil
			}
			if !info.IsDir() && !strings.HasSuffix(info.Name(), "_test.go") &&
				strings.HasSuffix(info.Name(), ".go") {
				out <- path
			}
			return nil
		})
		if err != nil {
			os.Exit(1)
		}
	}()

	return out
}
