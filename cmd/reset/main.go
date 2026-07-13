package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/tools/go/packages"
)

const tpl = `

package {{.Name}}

{{range .Structs}}

	func (v *{{.Name}}) Reset() {

		{{range .Fields}}

			{{if and .IsPremitive (not .IsStar)}}

				{{if eq .TypeName "string"}}

					v.{{.VarName}} = ""

				{{end}}

				{{if eq .TypeName "int"}}

					v.{{.VarName}} = 0

				{{end}}

				{{if eq .TypeName "bool"}}

					v.{{.VarName}} = false

				{{end}}

			{{end}}

			{{if and .IsPremitive .IsStar}}

				if v.{{.VarName}} != nil {

					{{if eq .TypeName "string"}}

						*v.{{.VarName}} = ""

					{{end}}

					{{if eq .TypeName "int"}}

						*v.{{.VarName}} = 0

					{{end}}

					{{if eq .TypeName "bool"}}

						*v.{{.VarName}} = false

					{{end}}

				}

			{{end}}

			{{if not .IsPremitive}}

				{{if .IsSlice}}

					v.{{.VarName}} = v.{{.VarName}}[:0]

				{{end}}

				{{if .IsMap}}

					clear(v.{{.VarName}})

				{{end}}

				{{if and .IsStruct (not .IsStar)}}

					v.{{.VarName}}.Reset()

				{{end}}

				{{if and .IsStruct .IsStar}}

					if v.{{.VarName}} != nil {
						v.{{.VarName}}.Reset()
					}

				{{end}}

			{{end}}

		{{end}}
	}

{{end}}
`

const (
	tplName  string = "gen"
	loadPath string = "./..."
)

var primitives = map[string]bool{
	"string":     true,
	"bool":       true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
	"byte":       true,
	"rune":       true,
}

type PkgData struct {
	Path    string
	Name    string
	Structs []St
}

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
	IsStruct    bool
}

func main() {
	// Конфигурация для инструмента загрузки всех пакетов
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Dir:  "",
	}

	// Сканирую все пакеты проекта в internal
	pkgs, err := packages.Load(cfg, loadPath)

	if err != nil {
		panic(err)
	}

	// Собираем мета данные
	allStructs := MakeMetaStructs(pkgs)

	// Формируем данные для шаблона
	genData := MakeTplData(pkgs, allStructs)

	if err := SaveFiles(genData); err != nil {
		panic(err)
	}

}

func SaveFiles(genData []PkgData) error {
	// Формирование шаблона
	t := template.Must(template.New(tplName).Parse(tpl))

	for _, pkg := range genData {

		var buf bytes.Buffer

		err := t.Execute(&buf, pkg)
		if err != nil {
			return err
		}

		bufFmt, err := format.Source(buf.Bytes())
		if err != nil {
			return err
		}

		err = os.WriteFile(
			filepath.Join(pkg.Path, "reset.gen.go"),
			bufFmt,
			0644,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func MakeTplData(pkgs []*packages.Package, allStructs map[string]bool) []PkgData {
	var result []PkgData

	for _, pkg := range pkgs {
		pkgData := PkgData{
			Name: pkg.Name,
		}

		for _, file := range pkg.Syntax {
			pkgData.Path = filepath.Dir(pkg.Fset.File(file.Pos()).Name())

			ast.Inspect(file, func(n ast.Node) bool {
				decl, ok := n.(*ast.GenDecl)
				if !ok {
					return true
				}

				if decl.Tok != token.TYPE {
					return true
				}

				for _, dec := range decl.Specs {
					tps, ok := dec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					myStruct, ok := tps.Type.(*ast.StructType)
					if !ok {
						continue
					}

					if decl.Doc == nil {
						continue
					}

					st := MakeData(decl, myStruct, tps, allStructs)
					if st != nil {
						pkgData.Structs = append(pkgData.Structs, *st)
					}
				}

				return true
			})
		}

		if len(pkgData.Structs) > 0 {
			result = append(result, pkgData)
		}
	}

	return result
}

// Формирование данных для генерации
func MakeData(
	decl *ast.GenDecl,
	myStruct *ast.StructType,
	tps *ast.TypeSpec,
	allStructs map[string]bool,
) *St {

	for _, comment := range decl.Doc.List {
		if comment.Text == "// generate:reset" {

			st, err := MakeStruct(myStruct, tps, allStructs)
			if err != nil {
				return nil
			}

			return st
		}
	}

	return nil
}

// Формирование структуры с полями для рендера
func MakeStruct(
	myStruct *ast.StructType,
	tps *ast.TypeSpec,
	allStructs map[string]bool,
) (*St, error) {

	var st St

	st.Name = tps.Name.Name

	if myStruct.Fields == nil {
		return nil, fmt.Errorf("no fields")
	}

	st.Fields = MakePayloads(myStruct, allStructs)

	return &st, nil
}

// Определяет является ли Ident примитивом или структурой
func IdentPayload(pl *Payload, allStructs map[string]bool, t *ast.Ident) {
	// Если это примитив
	if _, ok := primitives[t.Name]; ok {
		pl.IsPremitive = true
	}

	// Если это структура
	if _, ok := allStructs[t.Name]; ok {
		pl.IsStruct = true
	}

	pl.TypeName = t.Name
}

// Установка полезной нагрузки по всем филдам одной структуры
func MakePayloads(myStruct *ast.StructType, allStructs map[string]bool) []Payload {
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

		switch t := field.Type.(type) {
		case *ast.Ident:
			IdentPayload(&pl, allStructs, t)
			pls = append(pls, pl)
		case *ast.ArrayType:
			pl.IsSlice = true
			pls = append(pls, pl)
		case *ast.MapType:
			pl.IsMap = true
			pls = append(pls, pl)
		case *ast.StarExpr:
			// Для указателей
			switch starType := t.X.(type) {
			case *ast.Ident:
				IdentPayload(&pl, allStructs, starType)
				pl.IsStar = true
				pls = append(pls, pl)
			}
		}
	}

	return pls
}

// Сканирует проект, собирает структуры в множество. Это множество понадобится
// чтобы определить является ли Ident структурой
func MakeMetaStructs(pkgs []*packages.Package) map[string]bool {
	meta := make(map[string]bool)

	for _, pkg := range pkgs {
		// Перебираем все файлы пакета
		for _, file := range pkg.Syntax {
			pkg.Fset.File(file.Pos())
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

					// Наполняем мета данные названиями структур
					meta[tps.Name.Name] = true
				}

				return true
			})
		}
	}

	return meta
}
