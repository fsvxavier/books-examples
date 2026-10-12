package main

type ponto struct{ x, y int }

func novoPonto() *ponto { return &ponto{1, 2} } // escapa para o heap: o ponteiro sobrevive à função

func somaPontos() int {
	p := ponto{3, 4} // fica na pilha
	return p.x + p.y
}
