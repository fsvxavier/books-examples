package notif

import (
	"context"
	"errors"
	"fmt"
)

// Canal e a abstracao que o servico de notificacoes PRECISA (DIP, ISP: um metodo).
type Canal interface {
	Enviar(ctx context.Context, destino, msg string) error
}

// CanalFunc deixa qualquer funcao ser um Canal (OCP: novo canal = nova funcao).
type CanalFunc func(ctx context.Context, destino, msg string) error

func (f CanalFunc) Enviar(ctx context.Context, destino, msg string) error {
	return f(ctx, destino, msg)
}

var ErrSemCanal = errors.New("nenhum canal conseguiu enviar")

// Servico tenta os canais na ordem de preferencia, ate um funcionar (SRP: so orquestra).
type Servico struct {
	canais []nomeado
}

type nomeado struct {
	nome  string
	canal Canal
}

func NovoServico() *Servico { return &Servico{} }

func (s *Servico) Adicionar(nome string, c Canal) *Servico {
	s.canais = append(s.canais, nomeado{nome, c})
	return s
}

// Notificar devolve o nome do canal que funcionou e os erros dos que falharam antes.
func (s *Servico) Notificar(ctx context.Context, destino, msg string) (string, error) {
	var falhas []error
	for _, c := range s.canais {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		err := c.canal.Enviar(ctx, destino, msg)
		if err == nil {
			return c.nome, errors.Join(falhas...)
		}
		falhas = append(falhas, fmt.Errorf("%s: %w", c.nome, err))
	}
	return "", fmt.Errorf("%w: %w", ErrSemCanal, errors.Join(falhas...))
}

// ComTentativas e um decorator: repete o envio n vezes (extensao sem tocar no canal).
func ComTentativas(c Canal, n int) Canal {
	return CanalFunc(func(ctx context.Context, destino, msg string) error {
		var err error
		for range n {
			if err = c.Enviar(ctx, destino, msg); err == nil {
				return nil
			}
		}
		return fmt.Errorf("apos %d tentativas: %w", n, err)
	})
}
