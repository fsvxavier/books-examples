package main

import "fmt"

// Dimensao 1 (implementacao): cores.
type Cor interface{ Nome() string }

type Vermelho struct{}
type Azul struct{}

func (Vermelho) Nome() string { return "vermelho" }
func (Azul) Nome() string     { return "azul" }

// Dimensao 2 (abstracao): formas. Cada forma guarda uma Cor: esta e a "ponte".
type Forma interface{ Desenhar() string }

type Circulo struct {
	Raio float64
	Cor  Cor
}

type Quadrado struct {
	Lado float64
	Cor  Cor
}

func (c Circulo) Desenhar() string  { return fmt.Sprintf("circulo %s r=%.1f", c.Cor.Nome(), c.Raio) }
func (q Quadrado) Desenhar() string { return fmt.Sprintf("quadrado %s l=%.1f", q.Cor.Nome(), q.Lado) }

func main() {
	for _, f := range []Forma{
		Circulo{2, Vermelho{}}, Circulo{2, Azul{}},
		Quadrado{3, Vermelho{}}, Quadrado{3, Azul{}},
	} {
		fmt.Println(f.Desenhar())
	}
}
