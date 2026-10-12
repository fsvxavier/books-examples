package main

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"golang.org/x/sync/errgroup"
)

// ErroPanic embrulha um panic ocorrido em uma tarefa, com a pilha.
type ErroPanic struct {
	Valor any
	Pilha []byte
}

func (e *ErroPanic) Error() string { return fmt.Sprintf("panic na tarefa: %v", e.Valor) }

// Mapear aplica f a cada item, com no máximo limite tarefas ao mesmo
// tempo, e devolve os resultados NA ORDEM dos itens. O primeiro erro
// cancela o contexto das demais e é o erro devolvido; um panic em f vira
// *ErroPanic em vez de derrubar o processo.
func Mapear[E, S any](ctx context.Context, itens []E, limite int,
	f func(ctx context.Context, i int, e E) (S, error)) ([]S, error) {

	pai := ctx
	g, ctx := errgroup.WithContext(ctx)
	if limite > 0 {
		g.SetLimit(limite)
	}
	saida := make([]S, len(itens)) // cada tarefa escreve só no seu índice

	for i, item := range itens {
		if ctx.Err() != nil {
			break // já falhou ou foi cancelado: não lança mais nada
		}
		g.Go(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = &ErroPanic{Valor: r, Pilha: debug.Stack()}
				}
			}()
			if err := ctx.Err(); err != nil {
				return err // o grupo cancelou enquanto esperávamos vaga
			}
			saida[i], err = f(ctx, i, item)
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	// Se o contexto de fora foi cancelado antes de qualquer tarefa começar,
	// nenhuma delas falhou e Wait devolve nil: sem esta checagem, o chamador
	// receberia um resultado de zeros como se fosse sucesso.
	if err := pai.Err(); err != nil {
		return nil, err
	}
	return saida, nil
}

var ErroDeNegocio = errors.New("erro de negócio")
