package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const limiteCorpo = 4 << 10

// servidor liga as rotas HTTP ao armazém. Relógio e gerador de código são
// campos para que os testes possam fixá-los.
type servidor struct {
	armazem    *armazem
	novoCodigo geradorCodigo
	agora      func() time.Time
	baseURL    string
	log        *slog.Logger
	mux        *http.ServeMux
}

func novoServidor(cfg config, log *slog.Logger) *servidor {
	s := &servidor{
		armazem:    novoArmazem(),
		novoCodigo: codigoAleatorio,
		agora:      time.Now,
		baseURL:    strings.TrimRight(cfg.baseURL, "/"),
		log:        log,
		mux:        http.NewServeMux(),
	}
	s.mux.HandleFunc("POST /links", s.criar)
	s.mux.HandleFunc("GET /links/{codigo}", s.consultar)
	s.mux.HandleFunc("GET /{codigo}", s.redirecionar)
	s.mux.HandleFunc("GET /saude", s.saude)
	return s
}

func (s *servidor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type pedidoCriar struct {
	URL string `json:"url"`
}

type respostaCriar struct {
	Codigo string `json:"codigo"`
	Curto  string `json:"curto"`
}

func (s *servidor) criar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, limiteCorpo)
	var p pedidoCriar
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
		l := link{Codigo: s.novoCodigo(), Destino: destino, Criado: s.agora()}
		err := s.armazem.salvar(l)
		if errors.Is(err, errCodigoEmUso) {
			continue
		}
		if err != nil {
			s.erro(w, http.StatusInternalServerError, "falha ao salvar")
			return
		}
		s.log.Info("link criado", "codigo", l.Codigo)
		s.json(w, http.StatusCreated, respostaCriar{Codigo: l.Codigo, Curto: s.baseURL + "/" + l.Codigo})
		return
	}
	s.erro(w, http.StatusServiceUnavailable, "não consegui gerar um código livre")
}

func (s *servidor) redirecionar(w http.ResponseWriter, r *http.Request) {
	l, err := s.armazem.registrarAcesso(r.PathValue("codigo"))
	if errors.Is(err, errNaoEncontrado) {
		s.erro(w, http.StatusNotFound, err.Error())
		return
	}
	http.Redirect(w, r, l.Destino, http.StatusFound)
}

func (s *servidor) consultar(w http.ResponseWriter, r *http.Request) {
	l, err := s.armazem.buscar(r.PathValue("codigo"))
	if errors.Is(err, errNaoEncontrado) {
		s.erro(w, http.StatusNotFound, err.Error())
		return
	}
	s.json(w, http.StatusOK, l)
}

func (s *servidor) saude(w http.ResponseWriter, _ *http.Request) {
	s.json(w, http.StatusOK, map[string]string{"estado": "ok"})
}

func (s *servidor) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.log.Error("escrevendo resposta", "erro", err)
	}
}

func (s *servidor) erro(w http.ResponseWriter, status int, msg string) {
	s.json(w, status, map[string]string{"erro": msg})
}
