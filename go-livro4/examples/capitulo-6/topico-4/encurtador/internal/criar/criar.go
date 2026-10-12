// Package criar é a fatia "criar um link": a rota HTTP, a regra de tentativas
// diante de colisões, o SQL e a publicação do evento, tudo no mesmo lugar. Só
// conhece o que é realmente compartilhado: o domínio e o contrato.
package criar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/contrato"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/dominio"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/infra/resposta"
)

const (
	tentativasMax   = 5
	limiteCorpo     = 4 << 10
	codigoUnicidade = "23505" // SQLSTATE de violação de chave única
)

// Publicador é o que esta fatia exige de quem entrega eventos. A interface é
// local: a fatia não depende de nenhum pacote de infraestrutura para declará-la.
type Publicador interface {
	Publicar(ctx context.Context, eventos ...dominio.Evento)
}

// Servico é a fatia.
type Servico struct {
	pool     *pgxpool.Pool
	publicar Publicador
	gerar    func() string
	agora    func() time.Time
	validade time.Duration
}

// Novo monta a fatia. validade zero cria links sem prazo.
func Novo(pool *pgxpool.Pool, publicar Publicador, gerar func() string, agora func() time.Time, validade time.Duration) *Servico {
	return &Servico{pool: pool, publicar: publicar, gerar: gerar, agora: agora, validade: validade}
}

// Executar encurta um destino. O destino é validado uma vez; depois, a fatia
// sorteia códigos até achar um livre.
func (s *Servico) Executar(ctx context.Context, destino string) (dominio.Link, error) {
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
		// Os eventos são recolhidos ANTES de salvar, e só publicados depois do sucesso.
		eventos := l.RecolherEventos()
		err = s.salvar(ctx, l)
		switch {
		case err == nil:
			s.publicar.Publicar(ctx, eventos...)
			return l, nil
		case errors.Is(err, dominio.ErrCodigoEmUso):
			continue
		default:
			return dominio.Link{}, fmt.Errorf("salvando link: %w", err)
		}
	}
	return dominio.Link{}, dominio.ErrSemCodigoLivre
}

// salvar é o acesso a dados desta fatia: um INSERT, e a unicidade fica por
// conta do banco, que decide de forma atômica.
func (s *Servico) salvar(ctx context.Context, l dominio.Link) error {
	var expira *time.Time // NULL quando o link não tem prazo
	if fim, tem := l.Expira(); tem {
		expira = &fim
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO links (codigo, destino, criado, expira) VALUES ($1, $2, $3, $4)`,
		l.Codigo().String(), l.Destino().String(), l.Criado(), expira)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoUnicidade {
		return dominio.ErrCodigoEmUso
	}
	return err
}

// Registrar liga a rota POST /links ao mux. baseURL é o prefixo dos links curtos.
func (s *Servico) Registrar(mux *http.ServeMux, baseURL string, log *slog.Logger) {
	baseURL = strings.TrimRight(baseURL, "/")
	mux.HandleFunc("POST /links", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limiteCorpo)
		var pedido contrato.PedidoCriar
		if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
			resposta.Erro(w, log, http.StatusBadRequest, "corpo inválido: "+err.Error())
			return
		}
		l, err := s.Executar(r.Context(), pedido.URL)
		switch {
		case errors.Is(err, dominio.ErrDestinoInvalido):
			resposta.Erro(w, log, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, dominio.ErrSemCodigoLivre):
			resposta.Erro(w, log, http.StatusServiceUnavailable, err.Error())
		case err != nil:
			log.Error("falha inesperada ao criar", "erro", err)
			resposta.Erro(w, log, http.StatusInternalServerError, "erro interno")
		default:
			codigo := l.Codigo().String()
			resposta.JSON(w, log, http.StatusCreated, contrato.RespostaCriar{Codigo: codigo, Curto: baseURL + "/" + codigo})
		}
	})
}
