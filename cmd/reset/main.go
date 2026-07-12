package main

import (
	"fmt"
	"go/ast"
	"log"

	"golang.org/x/tools/go/packages"
)

type St struct {
	Name   string
	Fields []Payload
}

type Payload struct {
	VarName     string
	IsPremitive bool
	TypeName    string
	IsStar      bool
	IsSlice     bool
	IsMap       bool
	HasReset    bool
}

func main() {
	// Здесь будет слайс с данными для генерации
	var genData []St

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
					myStruct, ok := tps.Type.(*ast.StructType)

					if !ok {
						continue
					}

					// Если комментария нет, переходим к следующей структуре
					if decl.Doc == nil {
						continue
					}

					// ast.Print(pkg.Fset, file)

					// Перебираю комментарии структуры
					for _, comment := range decl.Doc.List {
						// Если в комментарии есть строка генерации это то что мне нужно
						if comment.Text == "// generate:reset" {
							log.Println(pkg.Name)
							log.Println(f.Name())
							log.Println(comment.Text)

							// Создаем структуру для генерации
							var structItem St
							// Задаем ей имя
							structItem.Name = tps.Name.String()

							// Если у структуры нету полей, ничего не делаем
							if myStruct.Fields == nil {
								continue
							}

							// Поля структуры для генерации
							var pls []Payload

							// Перебираем поля структуры
							for _, field := range myStruct.Fields.List {
								// Для кажого поля своя поезная нагрузка
								var pl Payload
								// устанавливаем имя поля
								// Пока у нас ограничение на структуру с синтаксисом
								// каждого поле на новой строке, без запятой
								pl.VarName = field.Names[0].String()
								fmt.Println(field.Type)

								// log.Println(field.Type)
								switch t := field.Type.(type) {
								case *ast.Ident: // Для примитивов
									pl.IsPremitive = true
									pl.TypeName = t.Name

									pls = append(pls, pl)
								case *ast.StarExpr:
									// Для указателей
									switch starType := t.X.(type) {
									case *ast.Ident: // Для примитивов с указателями
										pl.IsPremitive = true
										pl.TypeName = starType.Name
										pl.IsStar = true

										pls = append(pls, pl)
									}
								}
							}

							// myStruct.Fields.
							structItem.Fields = pls
							genData = append(genData, structItem)
						}
					}
				}

				return true
			})

		}

	}

	log.Println(genData)

}
