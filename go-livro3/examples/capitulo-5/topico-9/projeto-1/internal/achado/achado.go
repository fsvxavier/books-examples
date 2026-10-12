// Package achado é o vocabulário comum das etapas do portão: um Achado diz
// em que etapa, por qual regra, em que linha e com que mensagem algo foi recusado.
package achado

// Achado é um problema encontrado por uma etapa.
type Achado struct {
	Etapa string
	Regra string
	Linha int
	Msg   string
}
