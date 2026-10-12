// Package nucleo é o hexágono: o domínio (em nucleo/dominio), as portas e o
// serviço que as une. Declara as duas famílias de portas (entrada e saída) e
// não importa nada além do domínio e da biblioteca padrão. Não sabe que
// existem HTTP, terminal, PostgreSQL ou log.
package nucleo

import (
	"context"
	"errors"
	"fmt"

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// ErrSemCodigoLivre indica que todas as tentativas de sortear um código caíram
// em códigos já usados.
var ErrSemCodigoLivre = errors.New("não consegui gerar um código livre")

// tentativasMax é o número de sorteios antes de desistir.
const tentativasMax = 5

// Servico implementa a porta de entrada Encurtador usando as portas de saída.
type Servico struct {
	repo      Repositorio
	gerar     Gerador
	agora     Relogio
	auditoria Auditoria
}

var _ Encurtador = (*Servico)(nil)

// NovoServico monta o serviço com as suas dependências.
func NovoServico(repo Repositorio, gerar Gerador, agora Relogio, auditoria Auditoria) *Servico {
	return &Servico{repo: repo, gerar: gerar, agora: agora, auditoria: auditoria}
}

// Criar encurta um destino. Sorteia códigos até achar um livre.
func (s *Servico) Criar(ctx context.Context, destino string) (dominio.Link, error) {
	for range tentativasMax {
		l, err := dominio.NovoLink(s.gerar(), destino, s.agora())
		if err != nil {
			return dominio.Link{}, err
		}
		err = s.repo.Salvar(ctx, l)
		switch {
		case err == nil:
			s.auditoria.LinkCriado(ctx, l)
			return l, nil
		case errors.Is(err, dominio.ErrCodigoEmUso):
			continue
		default:
			return dominio.Link{}, fmt.Errorf("salvando link: %w", err)
		}
	}
	return dominio.Link{}, ErrSemCodigoLivre
}

// Resolver devolve o link de um código e conta um acesso.
func (s *Servico) Resolver(ctx context.Context, codigo string) (dominio.Link, error) {
	l, err := s.repo.RegistrarAcesso(ctx, codigo)
	if err != nil {
		return dominio.Link{}, err
	}
	s.auditoria.LinkAcessado(ctx, l)
	return l, nil
}

// Consultar devolve o link de um código, sem contar acesso.
func (s *Servico) Consultar(ctx context.Context, codigo string) (dominio.Link, error) {
	return s.repo.Buscar(ctx, codigo)
}
