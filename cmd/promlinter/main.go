package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

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

var (
	MetricType = map[int32]string{
		0: "COUNTER",
		1: "GAUGE",
		2: "SUMMARY",
		3: "UNTYPED",
		4: "HISTOGRAM",
	}
	withVendor *bool
)

func init() {
	// To see the log position, added for debugging.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	app := kingpin.New(filepath.Base(os.Args[0]), help)
	app.Version("v0.0.3")
	app.HelpFlag.Short('h')

	listCmd := app.Command("list", "List metrics name.")
	listPaths := listCmd.Arg("files", "Files to parse metrics.").Strings()
	listStrict := listCmd.Flag("strict", "Strict mode. If true, linter will output more issues including parsing failures.").
		Default("false").Short('s').Bool()

	listPrintAddPos := listCmd.Flag("add-position", "Add metric position column when printing the result.").Default("false").Bool()
	listPrintAddModule := listCmd.Flag("add-module", "Add metric module column when printing the result.").Default("false").Bool()

	listPrintAddHelp := listCmd.Flag("add-help", "Add metric help column when printing the result.").Default("false").Bool()
	listPrintFormat := listCmd.Flag("output", "Print result formatted as JSON/YAML/Markdown").Short('o').Enum("yaml", "json", "md")

	withVendor = listCmd.Flag("with-vendor", "Scan vendor packages.").Default("false").Bool()

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
		p := printer{
			fmt:         *listPrintFormat,
			addHelp:     *listPrintAddHelp,
			addPosition: *listPrintAddPos,
			addModule:   *listPrintAddModule,
			metrics:     metrics,
		}

		p.printMetrics()
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
			if withVendor != nil && !*withVendor &&
				(strings.HasPrefix(path, "vendor"+sep) || strings.Contains(path, sep+"vendor"+sep)) {
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

func (p *printer) printDefault() {
	tw := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	defer tw.Flush()

	fieldSep := "\t"
	if p.fmt == "md" {
		fieldSep = "|"
	}

	var (
		fields []string
	)

	if p.addPosition || p.addModule {
		fields = []string{
			"POSITION", "TYPE", "NAME", "LABELS",
		}
	} else {
		fields = []string{
			"TYPE", "NAME", "LABELS",
		}
	}

	if p.addHelp {
		fields = append(fields, "HELP")
	}

	if p.fmt == "md" {
		fmt.Fprintf(tw, "|%s|\n", strings.Join(fields, fieldSep))
	} else {
		fmt.Fprintf(tw, "%s\n", strings.Join(fields, fieldSep))
	}

	if p.fmt == "md" {
		fmt.Fprintf(tw, "%s|\n", strings.Repeat("|---", len(fields)))
	}

	for _, m := range p.metrics {

		help := "N/A"
		if m.MetricFamily.Help != nil {
			help = *m.MetricFamily.Help
		}

		labels := strings.Join(m.Labels(), ",")
		if labels == "" {
			labels = "N/A"
		}

		var lineArr []string

		mname := *m.MetricFamily.Name
		if p.fmt == "md" {
			mname = fmt.Sprintf("`%s`", *m.MetricFamily.Name)
			labels = fmt.Sprintf("`%s`", labels)
		}

		if (p.addPosition || p.addModule) && p.addHelp {
			lineArr = []string{
				p.pos(m.Pos.String()),
				MetricType[int32(*m.MetricFamily.Type)],
				mname,
				labels,
				help,
			}
		} else if p.addPosition || p.addModule {
			lineArr = []string{
				p.pos(m.Pos.String()),
				MetricType[int32(*m.MetricFamily.Type)],
				mname,
				labels,
			}

		} else if p.addHelp {
			lineArr = []string{
				MetricType[int32(*m.MetricFamily.Type)],
				mname,
				labels,
				help,
			}
		} else {
			lineArr = []string{
				MetricType[int32(*m.MetricFamily.Type)],
				mname,
				labels,
			}
		}

		if p.fmt == "md" {
			fmt.Fprintf(tw, "|%s|\n", strings.Join(lineArr, fieldSep))
		} else {
			fmt.Fprintf(tw, "%s\n", strings.Join(lineArr, fieldSep))
		}
	}
}

type printer struct {
	fmt                             string
	addHelp, addPosition, addModule bool
	metrics                         []promlinter.MetricFamilyWithPos
}

func (p *printer) pos(pos string) (x string) {
	if p.addModule {
		x = filepath.Dir(pos)
	} else {
		x = pos
	}

	if p.fmt == "md" {
		return fmt.Sprintf("*%s*", x) // italic file path
	}
	return
}

func (p *printer) printMetrics() {
	switch p.fmt {
	case "json":
		p.printAsJson()
		return
	case "yaml":
		p.printAsYaml()
		return
	default:
		p.printDefault()
		return
	}
}

func (p *printer) printAsYaml() {
	b, err := yaml.Marshal(toPrint(p.metrics))
	if err != nil {
		fmt.Printf("Failed: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(b))

}

func (p *printer) printAsJson() {
	b, err := json.MarshalIndent(toPrint(p.metrics), "", "  ")
	if err != nil {
		fmt.Printf("Failed: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(b))
}

type MetricForPrinting struct {
	Name     string
	Help     string
	Type     string
	Filename string
	Labels   []string
	Line     int
	Column   int
}

func toPrint(metrics []promlinter.MetricFamilyWithPos) []MetricForPrinting {
	p := []MetricForPrinting{}
	for _, m := range metrics {
		if m.MetricFamily != nil && *m.MetricFamily.Name != "" {
			if m.MetricFamily.Type == nil {
				continue
			}
			n := ""
			h := ""

			if m.MetricFamily.Name != nil {
				n = *m.MetricFamily.Name
			}
			if m.MetricFamily.Help != nil {
				h = *m.MetricFamily.Help
			}

			var labels []string
			for _, m := range m.MetricFamily.Metric {
				for idx, _ := range m.Label {
					if m.Label[idx].Name != nil {
						labels = append(labels, strings.Trim(*m.Label[idx].Name, `"`))
					}
				}
			}

			i := MetricForPrinting{
				Name:     n,
				Help:     h,
				Type:     MetricType[int32(*m.MetricFamily.Type)],
				Filename: m.Pos.Filename,
				Line:     m.Pos.Line,
				Column:   m.Pos.Column,
				Labels:   labels,
			}
			p = append(p, i)
		}
	}
	return p
}
