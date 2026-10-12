// Comando guardacheck reúne os analisadores do laboratório num único binário
// compatível com go vet -vettool.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"exemplo.com/guarda/internal/analisadores/errodescartado"
	"exemplo.com/guarda/internal/analisadores/semctx"
)

func main() {
	multichecker.Main(semctx.Analyzer, errodescartado.Analyzer)
}
