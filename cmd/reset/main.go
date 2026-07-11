package main

import (
	"go/ast"
	"log"

	"golang.org/x/tools/go/packages"
)

func main() {
	// Конфигурация для инструмента загрузки всех пакетов
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Dir:  "",
	}

	// Сканирую все пакеты проекта в internal
	pkgs, err := packages.Load(cfg, "./...")

	if err != nil {
		panic(err)
	}

	// Перебираем все пакеты проекта
	for _, pkg := range pkgs {
		// Перебираем все файлы пакета
		for _, file := range pkg.Syntax {
			f := pkg.Fset.File(file.Pos())
			// fmt.Println(f.Name())

			ast.Inspect(file, func(n ast.Node) bool {
				// Получаем только декларации
				decl, ok := n.(*ast.GenDecl)

				if !ok {
					return true
				}

				// Фильтруем на type декларацию
				if decl.Tok.String() != "type" {
					return true
				}

				// Перебираем все объявленные типы
				for _, dec := range decl.Specs {
					// Получаем тип спецификации
					tps, ok := dec.(*ast.TypeSpec)

					if !ok {
						continue
					}

					// Если тип структура - это то что мне нужно
					_, ok = tps.Type.(*ast.StructType)

					if !ok {
						continue
					}

					// Если комментария нет, переходим к следующей структуре
					if decl.Doc == nil {
						continue
					}

					// Перебираю комментарии структуры
					for _, comment := range decl.Doc.List {
						// Если в комментарии есть строка генерации это то что мне нужно
						if comment.Text == "// generate:reset" {
							log.Println(pkg.Name)
							log.Println(f.Name())
							log.Println(comment.Text)
						}
					}
				}

				return true
			})

		}

	}

}
