// Command e05 serve para inspecionar o código que o compilador gera para generics.
package main

import "fmt"

type valorer interface{ Valor() int }

type Inteiro int

func (i Inteiro) Valor() int { return int(i) }

type Outro int

func (o Outro) Valor() int { return int(o) * 2 }

type Texto string

func (t Texto) Valor() int { return len(t) }

type Caixa struct{ n int }

func (c *Caixa) Valor() int { return c.n }

type Pacote struct{ m int }

func (p *Pacote) Valor() int { return p.m }

//go:noinline
func Soma[T valorer](xs []T) int {
	t := 0
	for _, x := range xs {
		t += x.Valor()
	}
	return t
}

//go:noinline
func SomaDireta(xs []Inteiro) int {
	t := 0
	for _, x := range xs {
		t += int(x)
	}
	return t
}

func main() {
	fmt.Println(Soma([]Inteiro{1, 2}), Soma([]Outro{1, 2}), Soma([]Texto{"a", "bc"}))
	fmt.Println(Soma([]*Caixa{{1}, {2}}), Soma([]*Pacote{{3}}))
	fmt.Println(SomaDireta([]Inteiro{1, 2}))
}
