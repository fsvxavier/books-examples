package main

import "fmt"

// Sem Bridge: um tipo para CADA combinacao de forma e cor.
type Desenhavel interface{ Desenhar() string }

type CirculoVermelho struct{ Raio float64 }
type CirculoAzul struct{ Raio float64 }
type QuadradoVermelho struct{ Lado float64 }
type QuadradoAzul struct{ Lado float64 }

func (c CirculoVermelho) Desenhar() string  { return fmt.Sprintf("circulo vermelho r=%.1f", c.Raio) }
func (c CirculoAzul) Desenhar() string      { return fmt.Sprintf("circulo azul r=%.1f", c.Raio) }
func (q QuadradoVermelho) Desenhar() string { return fmt.Sprintf("quadrado vermelho l=%.1f", q.Lado) }
func (q QuadradoAzul) Desenhar() string     { return fmt.Sprintf("quadrado azul l=%.1f", q.Lado) }

func main() {
	for _, d := range []Desenhavel{CirculoVermelho{2}, CirculoAzul{2}, QuadradoVermelho{3}, QuadradoAzul{3}} {
		fmt.Println(d.Desenhar())
	}
}
