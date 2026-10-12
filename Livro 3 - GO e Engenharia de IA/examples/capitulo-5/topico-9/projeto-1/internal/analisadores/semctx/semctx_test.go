package semctx_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"exemplo.com/guarda/internal/analisadores/semctx"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), semctx.Analyzer, "a")
}
