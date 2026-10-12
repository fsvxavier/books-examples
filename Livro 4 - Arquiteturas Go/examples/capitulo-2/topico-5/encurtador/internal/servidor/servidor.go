package servidor

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"exemplos/livro4/capitulo-2/topico-5/encurtador/internal/armazem"
	"exemplos/livro4/capitulo-2/topico-5/encurtador/internal/contrato"
)

const limiteCorpo = 4 << 10

// Servidor liga as rotas HTTP ao armazém. Relógio e gerador de código são
// campos para que os testes possam fixá-los.
type Servidor struct {
	armazem    *armazem.Armazem
	novoCodigo geradorCodigo
	agora      func() time.Time
	baseURL    string
	log        *slog.Logger
	mux        *http.ServeMux
}

// Novo monta o servidor com o armazém recebido pronto.
func Novo(baseURL string, a *armazem.Armazem, log *slog.Logger) *Servidor {
	s := &Servidor{
		armazem:    a,
		novoCodigo: codigoAleatorio,
		agora:      time.Now,
		baseURL:    strings.TrimRight(baseURL, "/"),
		log:        log,
		mux:        http.NewServeMux(),
	}
	s.mux.HandleFunc("POST /links", s.criar)
	s.mux.HandleFunc("GET /links/{codigo}", s.consultar)
	s.mux.HandleFunc("GET /{codigo}", s.redirecionar)
	s.mux.HandleFunc("GET /saude", s.saude)
	return s
}

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
	destino, err := validarDestino(p.URL)
	if err != nil {
		s.erro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	for range tentativasMax {
		l := armazem.Link{Codigo: s.novoCodigo(), Destino: destino, Criado: s.agora()}
		err := s.armazem.Salvar(l)
		if errors.Is(err, armazem.ErrCodigoEmUso) {
			continue
		}
		if err != nil {
			s.erro(w, http.StatusInternalServerError, "falha ao salvar")
			return
		}
		s.log.Info("link criado", "codigo", l.Codigo)
		s.json(w, http.StatusCreated, contrato.RespostaCriar{Codigo: l.Codigo, Curto: s.baseURL + "/" + l.Codigo})
		return
	}
	s.erro(w, http.StatusServiceUnavailable, "não consegui gerar um código livre")
}

func (s *Servidor) redirecionar(w http.ResponseWriter, r *http.Request) {
	l, err := s.armazem.RegistrarAcesso(r.PathValue("codigo"))
	if errors.Is(err, armazem.ErrNaoEncontrado) {
		s.erro(w, http.StatusNotFound, err.Error())
		return
	}
	http.Redirect(w, r, l.Destino, http.StatusFound)
}

func (s *Servidor) consultar(w http.ResponseWriter, r *http.Request) {
	l, err := s.armazem.Buscar(r.PathValue("codigo"))
	if errors.Is(err, armazem.ErrNaoEncontrado) {
		s.erro(w, http.StatusNotFound, err.Error())
		return
	}
	s.json(w, http.StatusOK, l)
}

func (s *Servidor) saude(w http.ResponseWriter, _ *http.Request) {
	s.json(w, http.StatusOK, map[string]string{"estado": "ok"})
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
