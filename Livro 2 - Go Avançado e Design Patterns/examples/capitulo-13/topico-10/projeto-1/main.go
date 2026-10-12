package main

// Acrescentamos um novo elemento, Porto, e o método VisitarPorto à interface.

type Porto struct{ Nome string }

type Visitante interface {
	VisitarCidade(string)
	VisitarPorto(*Porto)
}

type ExportadorXML struct{}

func (ExportadorXML) VisitarCidade(string) {}

// ExportadorXML não foi atualizado: esqueceu VisitarPorto.

func main() {
	var v Visitante = ExportadorXML{}
	_ = v
}
