package main

import "exemplos/livro2/capitulo-13/topico-4/projeto-1/molde"

type SemExecutar struct{ molde.Padrao }

func (SemExecutar) Preparar() error { return nil }

func main() {
	_ = molde.Rodar(SemExecutar{})
}
