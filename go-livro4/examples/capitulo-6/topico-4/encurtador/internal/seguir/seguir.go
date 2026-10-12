// Package seguir é a fatia "seguir um link": a rota GET /{codigo}, a leitura do
// link, a regra da expiração (que mora no agregado) e a publicação do acesso.
package seguir

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/dominio"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/infra/resposta"
)

// Publicador é o que esta fatia exige de quem entrega eventos.
type Publicador interface {
	Publicar(ctx context.Context, eventos ...dominio.Evento)
}

// Servico é a fatia.
type Servico struct {
	pool     *pgxpool.Pool
	publicar Publicador
	agora    func() time.Time
}

// Novo monta a fatia.
func Novo(pool *pgxpool.Pool, publicar Publicador, agora func() time.Time) *Servico {
	return &Servico{pool: pool, publicar: publicar, agora: agora}
}

// Executar devolve o link, se ele ainda vale, e publica o evento do acesso. A
// contagem acontece em quem reage ao evento (a fatia contagem), e não aqui.
func (s *Servico) Executar(ctx context.Context, codigo string) (dominio.Link, error) {
	c, err := dominio.NovoCodigo(codigo)
	if err != nil {
		return dominio.Link{}, err
	}
	l, err := s.buscar(ctx, c)
	if err != nil {
		return dominio.Link{}, err
	}
	if err := l.Acessar(s.agora()); err != nil {
		return dominio.Link{}, err
	}
	s.publicar.Publicar(ctx, l.RecolherEventos()...)
	return l, nil
}

// buscar lê o link e o reconstitui: os dados voltam pelas regras do domínio.
func (s *Servico) buscar(ctx context.Context, codigo dominio.Codigo) (dominio.Link, error) {
	var cod, destino string
	var criado time.Time
	var expira *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT codigo, destino, criado, expira FROM links WHERE codigo = $1`, codigo.String(),
	).Scan(&cod, &destino, &criado, &expira)
	if errors.Is(err, pgx.ErrNoRows) {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	if err != nil {
		return dominio.Link{}, err
	}
	var fim time.Time
	if expira != nil {
		fim = *expira
	}
	return dominio.Reconstituir(cod, destino, criado, fim)
}

// Registrar liga a rota GET /{codigo} ao mux.
func (s *Servico) Registrar(mux *http.ServeMux, log *slog.Logger) {
	mux.HandleFunc("GET /{codigo}", func(w http.ResponseWriter, r *http.Request) {
		l, err := s.Executar(r.Context(), r.PathValue("codigo"))
		switch {
		case errors.Is(err, dominio.ErrNaoEncontrado), errors.Is(err, dominio.ErrCodigoInvalido):
			resposta.Erro(w, log, http.StatusNotFound, dominio.ErrNaoEncontrado.Error())
		case errors.Is(err, dominio.ErrExpirado):
			resposta.Erro(w, log, http.StatusGone, err.Error())
		case err != nil:
			log.Error("falha inesperada ao seguir", "erro", err)
			resposta.Erro(w, log, http.StatusInternalServerError, "erro interno")
		default:
			http.Redirect(w, r, l.Destino().String(), http.StatusFound)
		}
	})
}
