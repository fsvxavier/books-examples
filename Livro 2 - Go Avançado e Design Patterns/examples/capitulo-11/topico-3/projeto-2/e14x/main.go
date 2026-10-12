package main

import "exemplos/livro2/capitulo-11/topico-3/projeto-2/doc"

type falso struct{}

func (falso) Descricao() string { return "falso" }

func main() {
	var s doc.Snapshot = falso{}
	_ = s
}
