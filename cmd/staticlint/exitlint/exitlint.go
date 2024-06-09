// Package exitlint определяет анализатор для поиска вызова (os.Exit)
//
//	все анализаторы класса SA (или укзанные в файл конфигурации)
//	и стилистические анализатор класса ST
//	кастомный анализатор поска os.Exit
//	конфигурирование задется файлом config.json (если файл не задан, то проводятся все проверки, заданные по умолчанию)
//
// запуск линтера
//
//	go run cmd/staticlint/staticlint.go ./...
//	staticlint.go ./...
package exitlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// ExitCheckAnalyzer определение анализатора поиска os.Exit
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name:     "exitcheck",
	Doc:      "check for os exit",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

// процедура запуска поиска необходимиого условия
func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if c, ok := node.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok {
					var ident string
					if v, ok := s.X.(*ast.Ident); ok {
						ident = v.Name
					}
					if ident == "os" && s.Sel.Name == "Exit" {
						pass.Reportf(s.X.Pos(), "OMG why do you use the os.Exit!")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
