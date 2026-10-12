package main

import (
	"errors"
	"fmt"
	"iter"
	"runtime/debug"
)

type No interface {
	Valor() int
}

type Folha int

func (f Folha) Valor() int { return int(f) }

type Grupo struct {
	nome   string
	filhos []No
}

func (g *Grupo) Valor() int {
	soma := 0
	for _, f := range g.filhos {
		soma += f.Valor()
	}
	return soma
}

// Versao ingenua: aceita qualquer filho, inclusive o proprio grupo.
func (g *Grupo) AdicionarIngenuo(n No) { g.filhos = append(g.filhos, n) }

var ErrCiclo = errors.New("ciclo na arvore")

// contem diz se alvo esta em algum lugar da subarvore de n.
func contem(n No, alvo *Grupo) bool {
	if g, ok := n.(*Grupo); ok {
		if g == alvo {
			return true
		}
		for _, f := range g.filhos {
			if contem(f, alvo) {
				return true
			}
		}
	}
	return false
}

// Adicionar recusa o filho se ele contem o proprio pai (formaria um ciclo).
func (g *Grupo) Adicionar(n No) error {
	if contem(n, g) {
		return fmt.Errorf("adicionar a %q: %w", g.nome, ErrCiclo)
	}
	g.filhos = append(g.filhos, n)
	return nil
}

// Folhas percorre a arvore em profundidade como um iterador (iter.Seq).
func (g *Grupo) Folhas() iter.Seq[int] {
	return func(yield func(int) bool) {
		g.percorrer(yield)
	}
}

func (g *Grupo) percorrer(yield func(int) bool) bool {
	for _, f := range g.filhos {
		switch v := f.(type) {
		case *Grupo:
			if !v.percorrer(yield) {
				return false
			}
		default:
			if !yield(f.Valor()) {
				return false
			}
		}
	}
	return true
}

func main() {
	debug.SetMaxStack(5 << 20) // limita a pilha a 5 MiB so para o experimento ser rapido

	a, b := &Grupo{nome: "a"}, &Grupo{nome: "b"}
	a.Adicionar(Folha(1))
	b.Adicionar(Folha(2))
	fmt.Println("a contem b:", a.Adicionar(b))
	fmt.Println("b contem a (deveria recusar):", b.Adicionar(a))
	fmt.Println("a contem a (deveria recusar):", a.Adicionar(a))

	for v := range a.Folhas() {
		fmt.Println("folha:", v)
	}

	fmt.Println("agora a versao ingenua, que cria o ciclo:")
	c := &Grupo{nome: "c"}
	c.AdicionarIngenuo(Folha(1))
	c.AdicionarIngenuo(c)
	fmt.Println(c.Valor())
}
