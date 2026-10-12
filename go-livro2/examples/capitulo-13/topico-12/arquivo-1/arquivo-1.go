package main

import (
	"fmt"
	"strings"
)

type No interface{ isNo() }

type Cidade struct {
	Nome       string
	Habitantes int
}
type Industria struct {
	Nome         string
	Funcionarios int
}

func (*Cidade) isNo()    {}
func (*Industria) isNo() {}

// VisitanteR devolve um resultado do tipo R: sem campo acumulador.
type VisitanteR[R any] interface {
	VisitarCidade(*Cidade) R
	VisitarIndustria(*Industria) R
}

// Aceitar é uma FUNÇÃO genérica (métodos não podem ter parâmetros de
// tipo). Faz o despacho com um type switch e devolve o R do visitante.
func Aceitar[R any](n No, v VisitanteR[R]) R {
	switch n := n.(type) {
	case *Cidade:
		return v.VisitarCidade(n)
	case *Industria:
		return v.VisitarIndustria(n)
	}
	panic(fmt.Sprintf("Aceitar: tipo %T não tratado", n))
}

// Um visitante que devolve string...
type Nome struct{}

func (Nome) VisitarCidade(c *Cidade) string       { return "cidade " + c.Nome }
func (Nome) VisitarIndustria(i *Industria) string { return "indústria " + i.Nome }

// ...e outro que devolve int.
type Pessoas struct{}

func (Pessoas) VisitarCidade(c *Cidade) int       { return c.Habitantes }
func (Pessoas) VisitarIndustria(i *Industria) int { return i.Funcionarios }

// Grafo é um tipo CONCRETO. Desde o Go 1.27, os seus métodos podem ter
// parâmetros de tipo próprios, então Mapear é método, e não função solta.
type Grafo struct{ Nos []No }

func (g Grafo) Mapear[R any](v VisitanteR[R]) []R {
	out := make([]R, 0, len(g.Nos))
	for _, n := range g.Nos {
		out = append(out, Aceitar(n, v))
	}
	return out
}

func main() {
	g := Grafo{[]No{&Cidade{"Campinas", 1200000}, &Industria{"Alfa", 800}, &Cidade{"Jundiaí", 450000}}}

	nomes := g.Mapear(Nome{})
	fmt.Println(strings.Join(nomes, "; "))

	total := 0
	for _, p := range g.Mapear(Pessoas{}) {
		total += p
	}
	fmt.Println("pessoas:", total)
}
