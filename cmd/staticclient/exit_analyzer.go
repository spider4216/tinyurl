package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check os exit in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch m := node.(type) {
			case *ast.FuncDecl:
				if m.Name.String() == "main" {
					ast.Inspect(m, func(n ast.Node) bool {
						switch x := n.(type) {
						case *ast.CallExpr:
							if call, ok := x.Fun.(*ast.SelectorExpr); ok {
								if ident, ok := call.X.(*ast.Ident); ok {
									if ident.Name == "os" && call.Sel.String() == "Exit" {
										pass.Reportf(ident.NamePos, "os.Exit not allowed")
									}
								}
							}
						}

						return true
					})
				}
			}

			return true
		})
	}

	return nil, nil
}
