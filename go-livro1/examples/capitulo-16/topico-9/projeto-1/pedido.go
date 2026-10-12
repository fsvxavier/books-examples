// Package pedido tem um serviço simples para demonstrar testify e mockery.
package pedido

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNaoEncontrado = errors.New("pedido não encontrado")
	ErrValorInvalido = errors.New("valor deve ser positivo")
)

// Pedido representa uma compra. Valores em centavos.
type Pedido struct {
	ID    int
	Valor int64
	Pago  bool
}

// Repositorio guarda e busca pedidos.
type Repositorio interface {
	Buscar(ctx context.Context, id int) (Pedido, error)
	Salvar(ctx context.Context, p Pedido) error
}

// Notificador avisa o cliente sobre eventos do pedido.
type Notificador interface {
	Avisar(ctx context.Context, id int, mensagem string) error
}

// Servico reúne as regras de negócio.
type Servico struct {
	repo Repositorio
	nota Notificador
}

func NovoServico(repo Repositorio, nota Notificador) *Servico {
	return &Servico{repo: repo, nota: nota}
}

// Pagar marca o pedido como pago e avisa o cliente.
func (s *Servico) Pagar(ctx context.Context, id int) error {
	p, err := s.repo.Buscar(ctx, id)
	if err != nil {
		return fmt.Errorf("buscar pedido %d: %w", id, err)
	}
	if p.Valor <= 0 {
		return ErrValorInvalido
	}
	p.Pago = true
	if err := s.repo.Salvar(ctx, p); err != nil {
		return fmt.Errorf("salvar pedido %d: %w", id, err)
	}
	// a notificação é "melhor esforço": a falha não desfaz o pagamento
	_ = s.nota.Avisar(ctx, id, "pagamento confirmado")
	return nil
}
