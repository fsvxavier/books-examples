// Package nucleo é o hexágono: o domínio (em nucleo/dominio), as portas e o
// serviço que as une. Não importa nada além do domínio e da biblioteca padrão.
package nucleo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"exemplos/livro4/capitulo-5/topico-3/encurtador/internal/nucleo/dominio"
)

// ErrSemCodigoLivre indica que todas as tentativas de sortear um código caíram
// em códigos já usados.
var ErrSemCodigoLivre = errors.New("não consegui gerar um código livre")

// tentativasMax é o número de sorteios antes de desistir.
const tentativasMax = 5

// Servico implementa a porta de entrada Encurtador usando as portas de saída.
type Servico struct {
	repo       Repositorio
	contador   Contador
	publicador Publicador
	gerar      Gerador
	agora      Relogio
	validade   time.Duration
}

var _ Encurtador = (*Servico)(nil)

// NovoServico monta o serviço. validade zero cria links sem prazo.
func NovoServico(repo Repositorio, contador Contador, publicador Publicador, gerar Gerador, agora Relogio, validade time.Duration) *Servico {
	return &Servico{repo: repo, contador: contador, publicador: publicador, gerar: gerar, agora: agora, validade: validade}
}

// Criar encurta um destino. O destino é validado uma vez; depois, o serviço
// sorteia códigos até achar um livre.
func (s *Servico) Criar(ctx context.Context, destino string) (dominio.Link, error) {
	d, err := dominio.NovoDestino(destino)
	if err != nil {
		return dominio.Link{}, err
	}
	for range tentativasMax {
		c, err := dominio.NovoCodigo(s.gerar())
		if err != nil {
			return dominio.Link{}, err
		}
		l, err := dominio.NovoLink(c, d, s.agora(), s.validade)
		if err != nil {
			return dominio.Link{}, err
		}
		// Os eventos são recolhidos ANTES de salvar: o que o repositório
		// guarda é o estado do agregado, e nunca os eventos pendentes.
		eventos := l.RecolherEventos()
		err = s.repo.Salvar(ctx, l)
		switch {
		case err == nil:
			s.publicador.Publicar(ctx, eventos...)
			return l, nil
		case errors.Is(err, dominio.ErrCodigoEmUso):
			continue
		default:
			return dominio.Link{}, fmt.Errorf("salvando link: %w", err)
		}
	}
	return dominio.Link{}, ErrSemCodigoLivre
}

// Resolver devolve o link, se ele ainda vale, e publica o evento do acesso. A
// contagem acontece em quem reage ao evento, e não aqui.
func (s *Servico) Resolver(ctx context.Context, codigo string) (dominio.Link, error) {
	l, err := s.buscar(ctx, codigo)
	if err != nil {
		return dominio.Link{}, err
	}
	if err := l.Acessar(s.agora()); err != nil {
		return dominio.Link{}, err
	}
	s.publicador.Publicar(ctx, l.RecolherEventos()...)
	return l, nil
}

// Consultar devolve o link e o total de acessos, sem contar acesso.
func (s *Servico) Consultar(ctx context.Context, codigo string) (Consulta, error) {
	l, err := s.buscar(ctx, codigo)
	if err != nil {
		return Consulta{}, err
	}
	total, err := s.contador.Total(ctx, l.Codigo())
	if err != nil {
		return Consulta{}, fmt.Errorf("contando acessos: %w", err)
	}
	return Consulta{Link: l, Acessos: total}, nil
}

func (s *Servico) buscar(ctx context.Context, codigo string) (dominio.Link, error) {
	c, err := dominio.NovoCodigo(codigo)
	if err != nil {
		return dominio.Link{}, err
	}
	return s.repo.Buscar(ctx, c)
}

// ContarAcessos devolve o manipulador que reage a LinkAcessado somando um ao
// contador. É uma reação do núcleo a um fato do domínio, entregue pelo
// publicador que a composição escolher.
func ContarAcessos(c Contador) func(context.Context, dominio.Evento) error {
	return func(ctx context.Context, e dominio.Evento) error {
		acesso, ok := e.(dominio.LinkAcessado)
		if !ok {
			return nil
		}
		return c.Incrementar(ctx, acesso.Codigo)
	}
}
