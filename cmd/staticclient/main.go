// multichecker подвергает анализу выбранные файлы проекта.
// Содержит следующие анализаторы:
//
// Стандартные
//   - appends.Analyzer       // проверяет некорректное использование append
//   - assign.Analyzer        // проверяет бесполезные присваивания
//   - bools.Analyzer         // проверяет подозрительные логические выражения
//   - copylock.Analyzer      // проверяет копирование значений, содержащих блокировки
//   - defers.Analyzer        // проверяет ошибки при использовании defer
//   - errorsas.Analyzer      // проверяет корректность вызовов errors.As
//   - httpresponse.Analyzer  // проверяет ошибки при использовании HTTP ответов
//   - ifaceassert.Analyzer   // проверяет невозможные типы интерфейсов
//   - loopclosure.Analyzer   // проверяет захват переменных цикла замыканиями
//   - lostcancel.Analyzer    // проверяет пропущенные функции отмены context.CancelFunc
//   - nilfunc.Analyzer       // проверяет бессмысленное сравнение функций с nil
//   - printf.Analyzer        // проверяет корректность форматирующих вызовов Printf
//   - scannererr.Analyzer    // проверяет пропущенную проверку ошибки Scanner.Err
//   - shadow.Analyzer        // проверяет затенение переменных
//   - sqlrowserr.Analyzer    // проверяет пропущенную проверку ошибки Rows.Err
//   - structtag.Analyzer     // проверяет корректность тегов структур
//   - unmarshal.Analyzer     // проверяет передачу некорректных значений в Unmarshal
//   - unreachable.Analyzer   // проверяет недостижимый код
//   - unusedresult.Analyzer  // проверяет игнорирование результатов некоторых функций
//
// Внешние
//   - nilerr.Analyzer        // проверяет возврат nil вместо обнаруженной ошибки
//   - durationcheck.Analyzer // проверяет подозрительные операции с time.Duration
//
// Мой анализатор
//   - OsExitAnalyzer         // проверяет на наличие вызова os.Exit в пакете main функции main
//
// Механизм запуска следующий:
//
//	go run ./cmd/staticclient ./...
//	make mcheck
package main

import (
	"github.com/charithe/durationcheck"
	"github.com/gostaticanalysis/nilerr"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/scannererr"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sqlrowserr"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

const (
	SALyzerPrefix string = "SA"
)

func main() {
	checkers := []*analysis.Analyzer{
		// Стандартные анализаторы
		appends.Analyzer,
		assign.Analyzer,
		bools.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		scannererr.Analyzer,
		shadow.Analyzer,
		sqlrowserr.Analyzer,
		structtag.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,

		// Внешние анализаторы
		nilerr.Analyzer,
		durationcheck.Analyzer,

		// Мой пакет
		OsExitAnalyzer,
	}

	STList := map[string]bool{
		"ST1005": true,
		"ST1008": true,
	}

	SList := map[string]bool{
		"S1002": true,
	}

	// Статические анализаторы класса SA
	for _, v := range staticcheck.Analyzers {
		checkers = append(checkers, v.Analyzer)
	}

	// Статические анализаторы класса ST
	for _, v := range stylecheck.Analyzers {
		// Добавляем указанные в справочнике анализаторы
		if STList[v.Analyzer.Name] {
			checkers = append(checkers, v.Analyzer)
		}
	}

	// Статические анализаторы класса S
	for _, v := range simple.Analyzers {
		// Добавляем указанные в справочнике анализаторы
		if SList[v.Analyzer.Name] {
			checkers = append(checkers, v.Analyzer)
		}
	}

	multichecker.Main(
		checkers...,
	)
}
