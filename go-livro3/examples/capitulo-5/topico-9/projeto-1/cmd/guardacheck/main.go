// Comando guardacheck reúne os analisadores do laboratório num único binário
// compatível com go vet -vettool.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-5/topico-9/projeto-1/internal/analisadores/errodescartado"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-5/topico-9/projeto-1/internal/analisadores/semctx"
)

func main() {
	multichecker.Main(semctx.Analyzer, errodescartado.Analyzer)
}
