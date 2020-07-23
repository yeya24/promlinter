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
	app := kingpin.New(filepath.Base(os.Args[0]), "Prometheus metrics linter tool for golang.")
	app.Version("v0.0.1")
	app.HelpFlag.Short('h')

	paths := app.Arg("files", "Files to lint.").Strings()
	strict := app.Flag("strict", "Strict mode. If true, linter will output more issues including parsing failures.").
		Default("false").Short('s').Bool()

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

	for _, iss := range promlinter.Run(fileSet, files, *strict) {
		fmt.Printf("%s %s %s %s\n", iss.Pos.Start, iss.Pos.End, iss.Metric, iss.Text)
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
