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

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "Prometheus metrics linter tool golang.")
	app.Version("v0.0.1")
	app.HelpFlag.Short('h')

	files := app.Arg("files", "").Strings()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var paths []string
	var fs []*ast.File
	fset := token.NewFileSet()

	for _, path := range *files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Exit(1)
		}
		for f := range findFiles(path) {
			file, err := parser.ParseFile(fset, f, nil, parser.AllErrors)
			if err != nil {
				os.Exit(1)
			}
			fs = append(fs, file)
			paths = append(paths, f)
		}
	}

	for i := range fs {
		issues := promlinter.Run(fs[i], fset)
		for _, iss := range issues {
			fmt.Printf("%s: %s\n", iss.Metric, iss.Text)
		}
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