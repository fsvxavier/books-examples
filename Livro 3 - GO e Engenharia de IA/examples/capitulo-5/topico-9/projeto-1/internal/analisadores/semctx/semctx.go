// Package semctx denuncia chamadas que existem em duas versões, uma sem e
// outra com context.Context. Modelos de linguagem copiam o exemplo mais
// antigo e mais comum, que quase sempre é a versão sem contexto.
package semctx

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

// Analyzer é o analisador semctx.
var Analyzer = &analysis.Analyzer{
	Name:     "semctx",
	Doc:      "recusa chamadas bloqueantes sem context.Context quando existe versão com contexto",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// alternativas mapeia o nome completo da função para o que usar no lugar.
var alternativas = map[string]string{
	"net/http.Get":                "http.NewRequestWithContext e Client.Do",
	"net/http.Head":               "http.NewRequestWithContext e Client.Do",
	"net/http.Post":               "http.NewRequestWithContext e Client.Do",
	"net/http.PostForm":           "http.NewRequestWithContext e Client.Do",
	"os/exec.Command":             "exec.CommandContext",
	"net.Dial":                    "(*net.Dialer).DialContext",
	"net.DialTimeout":             "(*net.Dialer).DialContext",
	"(*database/sql.DB).Query":    "QueryContext",
	"(*database/sql.DB).QueryRow": "QueryRowContext",
	"(*database/sql.DB).Exec":     "ExecContext",
	"(*database/sql.DB).Begin":    "BeginTx",
	"(*database/sql.DB).Ping":     "PingContext",
	"(*database/sql.DB).Prepare":  "PrepareContext",
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		fn := typeutil.StaticCallee(pass.TypesInfo, call)
		if fn == nil {
			return
		}
		if alt, ok := alternativas[fn.FullName()]; ok {
			pass.Reportf(call.Pos(), "%s não aceita context.Context; use %s", fn.FullName(), alt)
		}
	})
	return nil, nil
}
