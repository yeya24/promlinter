package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

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

var MetricType = map[int32]string{
	0: "COUNTER",
	1: "GAUGE",
	2: "SUMMARY",
	3: "UNTYPED",
	4: "HISTOGRAM",
}

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), help)
	app.Version("v0.0.2")
	app.HelpFlag.Short('h')

	listCmd := app.Command("list", "List metrics name.")
	listPaths := listCmd.Arg("files", "Files to parse metrics.").Strings()
	listStrict := listCmd.Flag("strict", "Strict mode. If true, linter will output more issues including parsing failures.").
		Default("false").Short('s').Bool()
	listPrintAddPos := listCmd.Flag("add-position", "Add metric position column when printing the result.").Default("false").Bool()
	listPrintAddHelp := listCmd.Flag("add-help", "Add metric help column when printing the result.").Default("false").Bool()

	lintCmd := app.Command("lint", "Lint metrics via promlint.")
	lintPaths := lintCmd.Arg("files", "Files to parse metrics.").Strings()
	lintStrict := lintCmd.Flag("strict", "Strict mode. If true, linter will output more issues including parsing failures.").
		Default("false").Short('s').Bool()
	disableLintFuncs := lintCmd.Flag("disable", "Disable lint functions (repeated)."+
		"Supported options: Help, Counter, MetricUnits, HistogramSummaryReserved, MetricTypeInName, "+
		"ReservedChars, CamelCase, UnitAbbreviations").Short('d').Enums(promlinter.LintFuncNames...)

	parsedCmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	fileSet := token.NewFileSet()

	res := 0
	switch parsedCmd {
	case listCmd.FullCommand():
		metrics := promlinter.RunList(fileSet, findFiles(*listPaths, fileSet), *listStrict)
		printMetrics(metrics, *listPrintAddPos, *listPrintAddHelp)
	case lintCmd.FullCommand():
		setting := promlinter.Setting{Strict: *lintStrict, DisabledLintFuncs: *disableLintFuncs}
		for _, iss := range promlinter.RunLint(fileSet, findFiles(*lintPaths, fileSet), setting) {
			res++
			fmt.Printf("%s %s %s\n", iss.Pos, iss.Metric, iss.Text)
		}
	}

	os.Exit(res)
}

func findFiles(paths []string, fileSet *token.FileSet) []*ast.File {
	var files []*ast.File
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Exit(1)
		}
		for f := range walkDir(path) {
			file, err := parser.ParseFile(fileSet, f, nil, parser.AllErrors)
			if err != nil {
				os.Exit(1)
			}
			files = append(files, file)
		}
	}
	return files
}

func walkDir(root string) chan string {
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

func printMetrics(metrics []promlinter.MetricFamilyWithPos, addPosition, addHelp bool) {
	tw := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	defer tw.Flush()

	var header string
	if addPosition {
		header = "POSITION\tTYPE\tNAME"
	} else {
		header = "TYPE\tNAME"
	}

	if addHelp {
		header += "\tHELP"
	}

	fmt.Fprintln(tw, header)

	for _, m := range metrics {
		if addPosition && addHelp {
			fmt.Fprintf(tw, "%v\t%v\t%v\t%v\n", m.Pos, MetricType[int32(*m.MetricFamily.Type)], *m.MetricFamily.Name, *m.MetricFamily.Help)
		} else if addPosition {
			fmt.Fprintf(tw, "%v\t%v\t%v\n", m.Pos, MetricType[int32(*m.MetricFamily.Type)], *m.MetricFamily.Name)
		} else if addHelp {
			fmt.Fprintf(tw, "%v\t%v\t%v\n", MetricType[int32(*m.MetricFamily.Type)], *m.MetricFamily.Name, *m.MetricFamily.Help)
		} else {
			fmt.Fprintf(tw, "%v\t%v\n", MetricType[int32(*m.MetricFamily.Type)], *m.MetricFamily.Name)
		}
	}
}
