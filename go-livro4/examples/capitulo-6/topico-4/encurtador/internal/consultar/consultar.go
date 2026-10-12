// Package consultar é a fatia "consultar um link": uma única consulta SQL que
// já devolve o que a resposta precisa. Não reconstitui o agregado: consultar é
// leitura pura, sem regra de negócio a aplicar, e passar pelo domínio seria
// trabalho sem retorno.
package consultar

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/contrato"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/dominio"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/infra/resposta"
)

// Servico é a fatia.
type Servico struct {
	pool *pgxpool.Pool
}

// Novo monta a fatia.
func Novo(pool *pgxpool.Pool) *Servico {
	return &Servico{pool: pool}
}

// Executar devolve os dados do link e o total de acessos, em uma só consulta.
// O LEFT JOIN faz um link que ninguém acessou ainda aparecer com zero.
func (s *Servico) Executar(ctx context.Context, codigo string) (contrato.Link, error) {
	c, err := dominio.NovoCodigo(codigo)
	if err != nil {
		return contrato.Link{}, err
	}
	var l contrato.Link
	var expira *time.Time
	err = s.pool.QueryRow(ctx,
		`SELECT l.codigo, l.destino, l.criado, l.expira, COALESCE(a.total, 0)
		   FROM links l LEFT JOIN acessos a ON a.codigo = l.codigo
		  WHERE l.codigo = $1`, c.String(),
	).Scan(&l.Codigo, &l.Destino, &l.Criado, &expira, &l.Acessos)
	if errors.Is(err, pgx.ErrNoRows) {
		return contrato.Link{}, dominio.ErrNaoEncontrado
	}
	if err != nil {
		return contrato.Link{}, err
	}
	l.Expira = expira
	return l, nil
}

// Registrar liga a rota GET /links/{codigo} ao mux.
func (s *Servico) Registrar(mux *http.ServeMux, log *slog.Logger) {
	mux.HandleFunc("GET /links/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		l, err := s.Executar(r.Context(), r.PathValue("codigo"))
		switch {
		case errors.Is(err, dominio.ErrNaoEncontrado), errors.Is(err, dominio.ErrCodigoInvalido):
			resposta.Erro(w, log, http.StatusNotFound, dominio.ErrNaoEncontrado.Error())
		case err != nil:
			log.Error("falha inesperada ao consultar", "erro", err)
			resposta.Erro(w, log, http.StatusInternalServerError, "erro interno")
		default:
			resposta.JSON(w, log, http.StatusOK, l)
		}
	})
}
