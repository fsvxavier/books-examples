// Package pedido reúne o domínio de pedidos: o tipo, as regras e o serviço.
package pedido

import (
	"context"
	"errors"
	"fmt"
)

// Status é a etapa do pedido.
type Status string

const (
	Criado Status = "criado"
	Pago   Status = "pago"
)

// Pedido representa uma compra. Valores em centavos.
type Pedido struct {
	ID     int    `json:"id"`
	Valor  int64  `json:"valor"`
	Status Status `json:"status"`
}

var (
	ErrNaoEncontrado = errors.New("pedido não encontrado")
	ErrValorInvalido = errors.New("valor deve ser positivo")
	ErrJaPago        = errors.New("pedido já está pago")
)

// Repositorio e Cobrador são definidos AQUI, junto de quem os usa (o Servico).
type Repositorio interface {
	Salvar(p Pedido) (Pedido, error)
	Buscar(id int) (Pedido, error)
}

type Cobrador interface {
	Cobrar(ctx context.Context, ref string, valor int64) error
}

// Servico orquestra as regras de negócio.
type Servico struct {
	repo     Repositorio
	cobrador Cobrador
}

func NovoServico(repo Repositorio, cobrador Cobrador) *Servico {
	return &Servico{repo: repo, cobrador: cobrador}
}

func (s *Servico) Criar(valor int64) (Pedido, error) {
	if valor <= 0 {
		return Pedido{}, ErrValorInvalido
	}
	return s.repo.Salvar(Pedido{Valor: valor, Status: Criado})
}

func (s *Servico) Buscar(id int) (Pedido, error) { return s.repo.Buscar(id) }

func (s *Servico) Pagar(ctx context.Context, id int) (Pedido, error) {
	p, err := s.repo.Buscar(id)
	if err != nil {
		return Pedido{}, err
	}
	if p.Status == Pago {
		return Pedido{}, ErrJaPago
	}
	if err := s.cobrador.Cobrar(ctx, fmt.Sprintf("pedido-%d", p.ID), p.Valor); err != nil {
		return Pedido{}, fmt.Errorf("cobrar pedido %d: %w", id, err)
	}
	p.Status = Pago
	return s.repo.Salvar(p)
}
