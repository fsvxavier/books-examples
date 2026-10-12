package cfg

import "fmt"

// Valor e inicializado na carga do pacote, antes de qualquer init().
var Valor = carregar()

func carregar() string {
	fmt.Println("cfg: variavel de pacote inicializada")
	return "producao"
}

func init() {
	fmt.Println("cfg: init()")
}
