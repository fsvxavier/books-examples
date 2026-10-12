// Package errodescartado denuncia erros jogados fora: com _ de propósito
// (x, _ := f()) ou por esquecimento (f() sozinho numa linha).
package errodescartado

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

// Analyzer é o analisador errodescartado.
var Analyzer = &analysis.Analyzer{
	Name:     "errodescartado",
	Doc:      "denuncia erros descartados com _ ou por esquecimento",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var tipoErro = types.Universe.Lookup("error").Type()

// isencoes são chamadas cujo erro é sempre nil ou não vale a pena checar.
var isencoes = map[string]bool{
	"(*strings.Builder).WriteString": true,
	"(*strings.Builder).WriteByte":   true,
	"(*strings.Builder).WriteRune":   true,
	"(*strings.Builder).Write":       true,
	"(*bytes.Buffer).Write":          true,
	"(*bytes.Buffer).WriteString":    true,
	"(*bytes.Buffer).WriteByte":      true,
	"(*bytes.Buffer).WriteRune":      true,
}

// nome devolve o nome completo da função chamada.
func nome(pass *analysis.Pass, call *ast.CallExpr) string {
	if fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func); ok {
		return fn.FullName()
	}
	return "uma chamada"
}

// isenta diz se o erro da chamada pode ser ignorado.
func isenta(pass *analysis.Pass, call *ast.CallExpr) bool {
	n := nome(pass, call)
	return isencoes[n] || strings.HasPrefix(n, "fmt.Print") || strings.HasPrefix(n, "fmt.Fprint")
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.AssignStmt)(nil), (*ast.ExprStmt)(nil)}, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.ExprStmt:
			call, ok := n.X.(*ast.CallExpr)
			if ok && devolveErro(pass, call, -1) && !isenta(pass, call) {
				pass.Reportf(call.Pos(), "o erro devolvido por %s é ignorado", nome(pass, call))
			}
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if id, ok := lhs.(*ast.Ident); !ok || id.Name != "_" {
					continue
				}
				var call *ast.CallExpr
				idx := -1
				if len(n.Rhs) == 1 {
					call, _ = n.Rhs[0].(*ast.CallExpr)
					idx = i
				} else if i < len(n.Rhs) {
					call, _ = n.Rhs[i].(*ast.CallExpr)
					idx = 0
				}
				if call != nil && devolveErro(pass, call, idx) && !isenta(pass, call) {
					pass.Reportf(lhs.Pos(), "o erro devolvido por %s é descartado com _", nome(pass, call))
				}
			}
		}
	})
	return nil, nil
}

// devolveErro diz se a posição idx do resultado é um error (idx -1: algum).
func devolveErro(pass *analysis.Pass, call *ast.CallExpr, idx int) bool {
	t := pass.TypesInfo.TypeOf(call)
	if t == nil {
		return false
	}
	if tup, ok := t.(*types.Tuple); ok {
		for i := 0; i < tup.Len(); i++ {
			if (idx == -1 || idx == i) && types.Identical(tup.At(i).Type(), tipoErro) {
				return true
			}
		}
		return false
	}
	return (idx == -1 || idx == 0) && types.Identical(t, tipoErro)
}
