package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	mainPack string = "main"
	testFile string = ".test"
)

// OsExitAnalyzer анализатор проверяющий наличие вызова os.Exit
// В пакете main функции main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check os exit in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Исключаем .test файлы
	if strings.HasSuffix(pass.Pkg.Path(), testFile) {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch m := node.(type) {
			case *ast.FuncDecl:
				if m.Name.String() == mainPack {
					ast.Inspect(m, func(n ast.Node) bool {
						switch x := n.(type) {
						case *ast.CallExpr:
							if call, ok := x.Fun.(*ast.SelectorExpr); ok {
								if ident, ok := call.X.(*ast.Ident); ok {
									ispec, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
									if ok {
										if ispec.Imported().Path() == "os" && call.Sel.String() == "Exit" {
											pass.Reportf(ident.NamePos, "os.Exit not allowed")
										}
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
