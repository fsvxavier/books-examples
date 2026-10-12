// Package dev mostra a desvirtualização estática, sem perfil.
package dev

// Forma é uma interface com um método.
type Forma interface{ Area() float64 }

// Quadrado implementa Forma.
type Quadrado struct{ L float64 }

// Area implementa Forma.
func (q Quadrado) Area() float64 { return q.L * q.L }

// Circulo implementa Forma.
type Circulo struct{ R float64 }

// Area implementa Forma.
func (c Circulo) Area() float64 { return 3.14159 * c.R * c.R }

// Conhecida: o compilador vê o tipo concreto na atribuição.
func Conhecida() float64 {
	var f Forma = Quadrado{L: 2}
	return f.Area()
}

// Fabrica devolve a interface; a chamada vem de outro lugar.
//
//go:noinline
func Fabrica(n int) Forma {
	if n%2 == 0 {
		return Quadrado{L: float64(n)}
	}
	return Circulo{R: float64(n)}
}

// Desconhecida: o tipo concreto só se sabe em tempo de execução.
func Desconhecida(n int) float64 {
	f := Fabrica(n)
	return f.Area()
}
