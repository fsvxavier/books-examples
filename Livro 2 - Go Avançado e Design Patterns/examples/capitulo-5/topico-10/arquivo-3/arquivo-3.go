package main

import "fmt"

// Dimensao 1 (implementacao): cores.
type Cor interface{ Nome() string }

type Vermelho struct{}
type Azul struct{}
type Verde struct{} // NOVO: so um tipo e um metodo

func (Vermelho) Nome() string { return "vermelho" }
func (Azul) Nome() string     { return "azul" }
func (Verde) Nome() string    { return "verde" }

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

// NOVO: so um tipo e um metodo.
type Triangulo struct {
	Base, Altura float64
	Cor          Cor
}

func (t Triangulo) Desenhar() string {
	return fmt.Sprintf("triangulo %s b=%.1f h=%.1f", t.Cor.Nome(), t.Base, t.Altura)
}

func (c Circulo) Desenhar() string  { return fmt.Sprintf("circulo %s r=%.1f", c.Cor.Nome(), c.Raio) }
func (q Quadrado) Desenhar() string { return fmt.Sprintf("quadrado %s l=%.1f", q.Cor.Nome(), q.Lado) }

func main() {
	for _, f := range []Forma{
		Circulo{2, Vermelho{}}, Circulo{2, Azul{}},
		Quadrado{3, Vermelho{}}, Quadrado{3, Azul{}},
		Triangulo{4, 2, Verde{}}, Triangulo{4, 2, Azul{}}, Circulo{1, Verde{}},
	} {
		fmt.Println(f.Desenhar())
	}
}
