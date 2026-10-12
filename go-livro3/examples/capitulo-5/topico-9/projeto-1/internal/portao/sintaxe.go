package portao

import (
	"context"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-5/topico-9/projeto-1/internal/achado"
)

// EtapaSintaxe recusa fonte que não é Go válido e normaliza o que é, com o
// mesmo formatador do gofmt. Nada do que vem depois vê texto fora do padrão.
var EtapaSintaxe = Etapa{Nome: "sintaxe", Rodar: func(_ context.Context, st *Estado) ([]achado.Achado, error) {
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "solucao.go", st.Fonte, parser.ParseComments|parser.SkipObjectResolution); err != nil {
		var as []achado.Achado
		if lista, ok := err.(scanner.ErrorList); ok {
			for i, e := range lista {
				if i == 5 {
					break
				}
				as = append(as, achado.Achado{Etapa: "sintaxe", Regra: "S01", Linha: e.Pos.Line, Msg: e.Msg})
			}
		} else {
			as = append(as, achado.Achado{Etapa: "sintaxe", Regra: "S01", Msg: err.Error()})
		}
		return as, nil
	}
	saida, err := format.Source([]byte(st.Fonte))
	if err != nil {
		return []achado.Achado{{Etapa: "sintaxe", Regra: "S02", Msg: err.Error()}}, nil
	}
	st.Fonte = string(saida)
	return nil, st.gravar()
}}
