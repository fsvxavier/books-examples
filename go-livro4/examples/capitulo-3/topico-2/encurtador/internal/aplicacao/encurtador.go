// Package aplicacao contém os casos de uso do encurtador: a orquestração das
// regras do domínio. Declara as interfaces (portas) de que precisa, e quem as
// implementa são os adaptadores. Depende só do domínio.
package aplicacao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/dominio"
)

// ErrSemCodigoLivre indica que todas as tentativas de sortear um código caíram
// em códigos já usados.
var ErrSemCodigoLivre = errors.New("não consegui gerar um código livre")

// tentativasMax é o número de sorteios antes de desistir.
const tentativasMax = 5

// Repositorio é a porta de saída: tudo que o caso de uso precisa guardar.
// Fica aqui, ao lado de quem a consome, e não ao lado de quem a implementa.
// Os erros devolvidos são os do domínio (ErrCodigoEmUso, ErrNaoEncontrado).
type Repositorio interface {
	// Salvar guarda o link, ou devolve dominio.ErrCodigoEmUso.
	Salvar(ctx context.Context, l dominio.Link) error
	// Buscar devolve o link, ou dominio.ErrNaoEncontrado.
	Buscar(ctx context.Context, codigo string) (dominio.Link, error)
	// RegistrarAcesso soma um acesso, de forma atômica, e devolve o link atualizado.
	RegistrarAcesso(ctx context.Context, codigo string) (dominio.Link, error)
}

// Gerador sorteia um código novo a cada chamada.
type Gerador func() string

// Relogio informa o instante atual.
type Relogio func() time.Time

// Servico reúne os casos de uso.
type Servico struct {
	repo  Repositorio
	gerar Gerador
	agora Relogio
}

// NovoServico monta o serviço com as suas dependências.
func NovoServico(repo Repositorio, gerar Gerador, agora Relogio) *Servico {
	return &Servico{repo: repo, gerar: gerar, agora: agora}
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
	return s.repo.RegistrarAcesso(ctx, codigo)
}

// Consultar devolve o link de um código, sem contar acesso.
func (s *Servico) Consultar(ctx context.Context, codigo string) (dominio.Link, error) {
	return s.repo.Buscar(ctx, codigo)
}
