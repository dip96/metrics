// Package staticlint предоставляет набор статических анализаторов для проверки
// качества кода на языке Go. Анализаторы могут быть использованы как отдельно,
// так и в составе multichecker.
package staticlint

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer представляет собой статический анализатор, который проверяет,
// что функция os.Exit не вызывается напрямую в функции main пакета main.
// Этот анализатор использует инспектор AST для обхода абстрактного
// синтаксического дерева и поиска запрещенных вызовов.
//
// Имя: noosexit
// Описание: Проверяет, что os.Exit не вызывается напрямую в функции main пакета main
// Зависимости: inspect.Analyzer
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "Проверяет, что os.Exit не вызывается напрямую в функции main пакета main",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

// run - это основная функция анализатора noosexit. Она принимает контекст
// анализаи возвращает результат анализа
//
// Алгоритм работы:
//  1. Получает инспектор AST из результатов анализатора inspect.
//  2. Устанавливает фильтр для узлов AST, чтобы обрабатывать только объявления
//     функций (ast.FuncDecl) и вызовы функций (ast.CallExpr).
//  3. Использует метод Preorder инспектора для обхода AST в прямом порядке.
//  4. Для каждого узла проверяет:
//     a. Если это объявление функции main в пакете main, то выполняет
//     дополнительный обход этой функции.
//     b. Внутри функции main ищет вызовы os.Exit.
//  5. Если найден прямой вызов os.Exit в функции main, то сообщает об ошибке,
//     используя pass.Reportf.
//
// Параметры:
//   - pass: *analysis.Pass - контекст анализа, предоставляемый фреймворком
//     go/analysis. Содержит информацию о пакете, типах, файлах и т.д.
//
// Возвращаемые значения:
//   - interface{}: результат анализа (в данном случае nil, так как
//     результат не используется).
//   - error: ошибка, если она возникла во время анализа (в данном случае
//     всегда nil, так как ошибки обрабатываются через pass.Report*).
func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncDecl:
			if n.Name.Name == "main" && pass.Pkg.Name() == "main" {
				ast.Inspect(n, func(node ast.Node) bool {
					if call, ok := node.(*ast.CallExpr); ok {
						if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
							if x, ok := fun.X.(*ast.Ident); ok && x.Name == "os" && fun.Sel.Name == "Exit" {
								pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещен")
							}
						}
					}
					return true
				})
			}
		}
	})

	return nil, nil
}
