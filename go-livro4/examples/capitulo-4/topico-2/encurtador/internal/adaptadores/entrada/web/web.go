// Package web é o adaptador de entrada HTTP: traduz requisições em chamadas
// aos casos de uso e os resultados (e erros) do domínio em respostas HTTP.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/contrato"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

const limiteCorpo = 4 << 10

// Casos é o que o adaptador exige da camada de aplicação. A interface mora
// aqui, em quem a consome, e *nucleo.Servico a satisfaz sem saber dela.
type Casos interface {
	Criar(ctx context.Context, destino string) (dominio.Link, error)
	Resolver(ctx context.Context, codigo string) (dominio.Link, error)
	Consultar(ctx context.Context, codigo string) (dominio.Link, error)
}

var _ Casos = (*nucleo.Servico)(nil)

// Servidor liga as rotas HTTP aos casos de uso.
type Servidor struct {
	casos   Casos
	baseURL string
	log     *slog.Logger
	mux     *http.ServeMux
}

// Novo monta o servidor. baseURL é o prefixo dos links curtos devolvidos.
func Novo(baseURL string, casos Casos, log *slog.Logger) *Servidor {
	s := &Servidor{
		casos:   casos,
		baseURL: strings.TrimRight(baseURL, "/"),
		log:     log,
		mux:     http.NewServeMux(),
	}
	s.mux.HandleFunc("POST /links", s.criar)
	s.mux.HandleFunc("GET /links/{codigo}", s.consultar)
	s.mux.HandleFunc("GET /{codigo}", s.redirecionar)
	s.mux.HandleFunc("GET /saude", s.saude)
	return s
}

// ServeHTTP faz do Servidor um http.Handler.
func (s *Servidor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Servidor) criar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, limiteCorpo)
	var p contrato.PedidoCriar
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		s.erro(w, http.StatusBadRequest, "corpo inválido: "+err.Error())
		return
	}
	l, err := s.casos.Criar(r.Context(), p.URL)
	if err != nil {
		s.falha(w, err)
		return
	}
	s.log.Info("link criado", "codigo", l.Codigo)
	s.json(w, http.StatusCreated, contrato.RespostaCriar{Codigo: l.Codigo, Curto: s.baseURL + "/" + l.Codigo})
}

func (s *Servidor) redirecionar(w http.ResponseWriter, r *http.Request) {
	l, err := s.casos.Resolver(r.Context(), r.PathValue("codigo"))
	if err != nil {
		s.falha(w, err)
		return
	}
	http.Redirect(w, r, l.Destino, http.StatusFound)
}

func (s *Servidor) consultar(w http.ResponseWriter, r *http.Request) {
	l, err := s.casos.Consultar(r.Context(), r.PathValue("codigo"))
	if err != nil {
		s.falha(w, err)
		return
	}
	s.json(w, http.StatusOK, contrato.Link{
		Codigo: l.Codigo, Destino: l.Destino, Criado: l.Criado, Acessos: l.Acessos,
	})
}

func (s *Servidor) saude(w http.ResponseWriter, _ *http.Request) {
	s.json(w, http.StatusOK, map[string]string{"estado": "ok"})
}

// falha traduz um erro do domínio ou da aplicação em status HTTP. Erros que o
// negócio não reconhece viram 500, sem vazar detalhes para o cliente.
func (s *Servidor) falha(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dominio.ErrDestinoInvalido):
		s.erro(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, dominio.ErrNaoEncontrado), errors.Is(err, dominio.ErrCodigoInvalido):
		s.erro(w, http.StatusNotFound, dominio.ErrNaoEncontrado.Error())
	case errors.Is(err, nucleo.ErrSemCodigoLivre):
		s.erro(w, http.StatusServiceUnavailable, err.Error())
	default:
		s.log.Error("falha inesperada", "erro", err)
		s.erro(w, http.StatusInternalServerError, "erro interno")
	}
}

func (s *Servidor) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.log.Error("escrevendo resposta", "erro", err)
	}
}

func (s *Servidor) erro(w http.ResponseWriter, status int, msg string) {
	s.json(w, status, contrato.Erro{Mensagem: msg})
}
