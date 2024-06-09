package pkg1

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func findExit() {
	src := `package main
	import (
		"fmt"
		"os"
	)
	func main() {
		fmt.Println("Hello, world!")
		os.Exit(1)
	}`
	fset := token.NewFileSet()
	// получаем дерево разбора
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok {
				if s.X.(*ast.Ident).Name == "os" && s.Sel.Name == "Exit" {
					return true //fmt.Sprintf("%v, %v", s.Sel.Name, fset.Position(s.Pos()))
				}
			}
		}
		return true
	})
}

func errCheckFunc() {
	res := findExit() // want "assignment with unchecked error"
	fmt.Println(res)  // want "expression returns unchecked error"
}
