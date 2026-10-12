// Package historico define o guardião (caretaker): empilha e devolve
// mementos sem saber o que há dentro deles.
package historico

import "errors"

// ErrVazio indica que não há mais o que desfazer.
var ErrVazio = errors.New("historico: nada para desfazer")

// Historico guarda mementos de qualquer tipo M. O guardião não exige
// nenhuma interface de M: ele só empilha e desempilha.
type Historico[M any] struct {
	pilha  []M
	limite int
}

// Novo cria um histórico que guarda no máximo limite mementos.
func Novo[M any](limite int) *Historico[M] {
	return &Historico[M]{limite: limite}
}

// Guardar empilha um memento e descarta o mais antigo se passar do limite.
func (h *Historico[M]) Guardar(m M) {
	h.pilha = append(h.pilha, m)
	if len(h.pilha) > h.limite {
		var zero M
		h.pilha[0] = zero // libera a referência para o coletor
		h.pilha = h.pilha[1:]
	}
}

// Desfazer desempilha o memento mais recente.
func (h *Historico[M]) Desfazer() (M, error) {
	var zero M
	if len(h.pilha) == 0 {
		return zero, ErrVazio
	}
	m := h.pilha[len(h.pilha)-1]
	h.pilha[len(h.pilha)-1] = zero
	h.pilha = h.pilha[:len(h.pilha)-1]
	return m, nil
}

// Tamanho informa quantos mementos estão guardados.
func (h *Historico[M]) Tamanho() int { return len(h.pilha) }
