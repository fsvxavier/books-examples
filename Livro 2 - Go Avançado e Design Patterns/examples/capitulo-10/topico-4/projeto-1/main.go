package main

import (
	"fmt"
	"iter"
	"slices"
)

func exemplo() *Arvore[int] {
	return &Arvore[int]{Raiz: &No[int]{1,
		&No[int]{2, &No[int]{Valor: 4}, &No[int]{Valor: 5}},
		&No[int]{3, nil, &No[int]{Valor: 6}},
	}}
}

func main() {
	a := exemplo()

	fmt.Print("profundidade:")
	for v := range a.All() {
		fmt.Print(" ", v)
	}
	fmt.Println()

	fmt.Print("largura:     ")
	for v := range a.Largura() {
		fmt.Print(" ", v)
	}
	fmt.Println()

	fmt.Println("folhas:      ", slices.Collect(a.Folhas()))

	fmt.Print("com nivel:   ")
	for nivel, v := range a.ComNivel() {
		fmt.Printf(" %d@%d", v, nivel)
	}
	fmt.Println()

	// Parada antecipada: break manda yield devolver false, e a travessia PARA de visitar.
	a.Visitas = 0
	for v := range a.All() {
		if v == 2 {
			break
		}
	}
	fmt.Println("nos visitados ate achar o 2 e dar break:", a.Visitas, "de 6")

	// Dois percursos simultaneos da mesma arvore: iter.Pull transforma em "puxar um por vez".
	proximoA, paraA := iter.Pull(a.All())
	defer paraA()
	proximoB, paraB := iter.Pull(a.Largura())
	defer paraB()
	fmt.Print("intercalados:")
	for {
		x, ok1 := proximoA()
		y, ok2 := proximoB()
		if !ok1 || !ok2 {
			break
		}
		fmt.Printf(" %d/%d", x, y)
	}
	fmt.Println()
}
