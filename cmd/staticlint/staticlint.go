package main

import (
	"encoding/json"
	"os"

	"github.com/closable/go-yandex-loyalty/cmd/staticlint/exitlint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Config — имя файла конфигурации.
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string
	Stylecheck  []string
}

func main() {
	var noconfig bool

	data, err := os.ReadFile(Config)
	if err != nil {
		noconfig = true
	}
	var cfg ConfigData
	mychecks := []*analysis.Analyzer{
		// кастомный анализатор но os.Exit
		exitlint.ExitCheckAnalyzer,
		// mismatches between assembly files and Go declarations
		asmdecl.Analyzer,
		// This checker reports assignments of the form x = x or a[i] = a[i].
		assign.Analyzer,
		// atomic: check for common mistakes using the sync/atomic package
		atomic.Analyzer,
		// Package bools defines an Analyzer that detects common mistakes involving boolean operators.
		bools.Analyzer,
		// Package buildtag defines an Analyzer that checks build tags.
		buildtag.Analyzer,
		// Package cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
		cgocall.Analyzer,
		// composites.Analyzer,
		// copylocks.Analyzer,
		// Package directive defines an Analyzer that checks known Go toolchain directives.
		directive.Analyzer,
		// he errorsas package defines an Analyzer that checks that the second argument to errors.As is a pointer to a type
		errorsas.Analyzer,
		// Package framepointer defines an Analyzer that reports assembly code that clobbers the frame pointer before saving it.
		framepointer.Analyzer,
		// Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
		httpresponse.Analyzer,
		// Package ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
		ifaceassert.Analyzer,
		// Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
		loopclosure.Analyzer,
		// Package lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
		lostcancel.Analyzer,
		// Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
		nilfunc.Analyzer,
		// Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
		printf.Analyzer,
		// Package shift defines an Analyzer that checks for shifts that exceed the width of an integer.
		shift.Analyzer,
		// Package shadow defines an Analyzer that checks for shadowed variables.
		shadow.Analyzer,
		// Package sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal
		sigchanyzer.Analyzer,
		// Package sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal
		slog.Analyzer,
		// stdmethods.Analiyer,
		// Package stringintconv defines an Analyzer that flags type conversions from integers to strings.
		stringintconv.Analyzer,
		// Package structtag defines an Analyzer that checks struct field tags are well formed.
		structtag.Analyzer,
		// Package testinggoroutine defines an Analyzerfor detecting calls to Fatal from a test goroutine.
		testinggoroutine.Analyzer,
		// Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
		tests.Analyzer,
		// Package timeformat defines an Analyzer that checks for the use of time.Format or time.Parse calls with a bad format.
		timeformat.Analyzer,
		// The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
		unmarshal.Analyzer,
		// Package unreachable defines an Analyzer that checks for unreachable code.
		unreachable.Analyzer,
		// Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
		unsafeptr.Analyzer,
		// Package unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
		unusedresult.Analyzer,
	}
	checks := make(map[string]bool)
	// если конфиг не найден, добавляем все
	if noconfig {
		for _, v := range staticcheck.Analyzers {
			mychecks = append(mychecks, v.Analyzer)
		}
		for _, v := range stylecheck.Analyzers {
			mychecks = append(mychecks, v.Analyzer)
		}

	} else {
		if err = json.Unmarshal(data, &cfg); err != nil {
			panic(err)
		}
		// staticcheck
		for _, v := range cfg.Staticcheck {
			checks[v] = true
		}
		// stylecheck
		for _, v := range cfg.Stylecheck {
			checks[v] = true
		}

		// добавляем анализаторы из staticcheck, которые указаны в файле конфигурации
		for _, v := range staticcheck.Analyzers {
			if checks[v.Analyzer.Name] {
				mychecks = append(mychecks, v.Analyzer)
			}
		}
		for _, v := range stylecheck.Analyzers {
			if checks[v.Analyzer.Name] {
				mychecks = append(mychecks, v.Analyzer)
			}
		}

	}

	multichecker.Main(
		mychecks...,
	)

}
