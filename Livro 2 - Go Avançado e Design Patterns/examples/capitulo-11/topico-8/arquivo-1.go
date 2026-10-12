package main

import (
	"fmt"
	"iter"
)

// Leitor percorre linhas e sabe guardar e retomar a posição: a Marca é o
// memento do estado da iteração.
type Leitor struct {
	linhas []string
	pos    int
}

// Marca é opaca para o cliente: só o Leitor sabe o que há dentro.
type Marca struct{ pos int }

func (l *Leitor) Marcar() Marca  { return Marca{l.pos} }
func (l *Leitor) Voltar(m Marca) { l.pos = m.pos }

// Restantes entrega as linhas a partir da posição atual e AVANÇA a posição
// a cada linha entregue; um break deixa o leitor exatamente onde parou.
func (l *Leitor) Restantes() iter.Seq[string] {
	return func(yield func(string) bool) {
		for l.pos < len(l.linhas) {
			linha := l.linhas[l.pos]
			l.pos++
			if !yield(linha) {
				return
			}
		}
	}
}

func main() {
	l := &Leitor{linhas: []string{"um", "dois", "três", "quatro", "cinco", "seis"}}

	for s := range l.Restantes() {
		fmt.Println("primeira passada:", s)
		if s == "dois" {
			break
		}
	}
	marca := l.Marcar()
	fmt.Println("marca guardada em", marca)

	for s := range l.Restantes() {
		fmt.Println("segunda passada: ", s)
		if s == "quatro" {
			break
		}
	}

	l.Voltar(marca)
	fmt.Println("voltou para a marca")
	for s := range l.Restantes() {
		fmt.Println("terceira passada:", s)
	}
}
