// Package e01 compara uma pilha de any com uma pilha genérica.
package e01

// PilhaAntiga guarda qualquer coisa: os valores passam por any.
type PilhaAntiga struct {
	elementos []any
}

// Empilhar guarda um valor.
func (p *PilhaAntiga) Empilhar(valor any) {
	p.elementos = append(p.elementos, valor)
}

// Desempilhar devolve o último valor, ou nil se a pilha está vazia.
func (p *PilhaAntiga) Desempilhar() any {
	if len(p.elementos) == 0 {
		return nil
	}
	ultimo := p.elementos[len(p.elementos)-1]
	p.elementos = p.elementos[:len(p.elementos)-1]
	return ultimo
}

// Pilha guarda valores de um só tipo T.
type Pilha[T any] struct {
	elementos []T
}

// Empilhar guarda um valor.
func (p *Pilha[T]) Empilhar(valor T) {
	p.elementos = append(p.elementos, valor)
}

// Desempilhar devolve o último valor e true; em uma pilha vazia, o zero de T e false.
func (p *Pilha[T]) Desempilhar() (T, bool) {
	var zero T
	if len(p.elementos) == 0 {
		return zero, false
	}
	ultimo := p.elementos[len(p.elementos)-1]
	p.elementos = p.elementos[:len(p.elementos)-1]
	return ultimo, true
}
