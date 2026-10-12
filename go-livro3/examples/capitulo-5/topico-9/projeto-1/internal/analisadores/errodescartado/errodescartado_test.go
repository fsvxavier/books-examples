package errodescartado_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"exemplo.com/guarda/internal/analisadores/errodescartado"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errodescartado.Analyzer, "a")
}
