package log

import (
	"fmt"

	"exemplos/livro2/capitulo-4/topico-6/projeto-1/cfg"
)

var Nivel = definirNivel()

func definirNivel() string {
	fmt.Println("log: variavel de pacote (usa cfg.Valor =", cfg.Valor+")")
	return "info"
}

func init() {
	fmt.Println("log: init()")
}
