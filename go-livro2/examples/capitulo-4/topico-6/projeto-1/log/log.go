package log

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-6/projeto-1/cfg"
)

var Nivel = definirNivel()

func definirNivel() string {
	fmt.Println("log: variavel de pacote (usa cfg.Valor =", cfg.Valor+")")
	return "info"
}

func init() {
	fmt.Println("log: init()")
}
