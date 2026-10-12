// app/main_test.go
package main

import (
	"testing"

	"exemplos/livro1/capitulo-14/topico-16/arquivo-2"
)

func TestIntegracao(t *testing.T) {
	if lib.Saudar("x") == "" {
		t.Error("vazio")
	}
}
