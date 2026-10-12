package main

import (
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// Num pacote único o compilador não impõe fronteiras. Este teste impõe: ele
// lê só os imports de cada arquivo e confere se a divisão por
// responsabilidade, feita pelos nomes dos arquivos, continua de pé.
func TestFronteirasEntreArquivos(t *testing.T) {
	proibidos := map[string][]string{
		"armazem.go":   {"net/http", "encoding/json", "log/slog"},
		"codigo.go":    {"net/http", "sync"},
		"validacao.go": {"net/http", "sync", "crypto/rand"},
		"servidor.go":  {"crypto/rand", "sync"},
	}
	fset := token.NewFileSet()
	for arquivo, lista := range proibidos {
		f, err := parser.ParseFile(fset, arquivo, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			caminho, _ := strconv.Unquote(imp.Path.Value)
			for _, p := range lista {
				if caminho == p {
					t.Errorf("%s não pode importar %q", arquivo, p)
				}
			}
		}
	}
}
