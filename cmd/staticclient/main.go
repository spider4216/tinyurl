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
