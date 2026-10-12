package errodescartado_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-5/topico-9/projeto-1/internal/analisadores/errodescartado"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errodescartado.Analyzer, "a")
}
