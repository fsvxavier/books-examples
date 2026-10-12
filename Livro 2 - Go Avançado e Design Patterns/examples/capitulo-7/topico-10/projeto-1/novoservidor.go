package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// ---------- Lado do servidor: expoe qualquer Perfis por HTTP ----------

func NovoServidor(p Perfis) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /perfis/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "id invalido", http.StatusBadRequest)
			return
		}
		perfil, err := p.Obter(contextoDe(r), id)
		if err != nil {
			erroHTTP(w, err)
			return
		}
		json.NewEncoder(w).Encode(perfil)
	})
	mux.HandleFunc("PUT /perfis/{id}", func(w http.ResponseWriter, r *http.Request) {
		var perfil Perfil
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&perfil); err != nil {
			http.Error(w, "corpo invalido", http.StatusBadRequest)
			return
		}
		if err := p.Atualizar(contextoDe(r), perfil); err != nil {
			erroHTTP(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}

// contextoDe recupera a identidade do cabecalho (so para o exemplo; em producao seria um token validado).
func contextoDe(r *http.Request) context.Context {
	ctx := r.Context()
	if v := r.Header.Get("X-Usuario"); v != "" {
		id, _ := strconv.Atoi(v)
		ctx = ComUsuario(ctx, Usuario{ID: id, Admin: r.Header.Get("X-Admin") == "1"})
	}
	return ctx
}

func erroHTTP(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNaoEncontrado):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrProibido):
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, "erro interno", http.StatusInternalServerError)
	}
}

// ---------- Proxy REMOTO: parece um Perfis local, mas fala HTTP ----------

type ClienteRemoto struct {
	base string
	http *http.Client
}

func NovoClienteRemoto(base string, c *http.Client) *ClienteRemoto {
	return &ClienteRemoto{base: base, http: c}
}

func (c *ClienteRemoto) fazer(ctx context.Context, metodo, caminho string, corpo []byte) (*http.Response, error) {
	var r io.Reader
	if corpo != nil {
		r = bytes.NewReader(corpo)
	}
	req, err := http.NewRequestWithContext(ctx, metodo, c.base+caminho, r)
	if err != nil {
		return nil, err
	}
	if u, ok := UsuarioDe(ctx); ok { // propaga a identidade
		req.Header.Set("X-Usuario", strconv.Itoa(u.ID))
		if u.Admin {
			req.Header.Set("X-Admin", "1")
		}
	}
	return c.http.Do(req)
}

// traduzir converte o status HTTP de volta nos erros do dominio: o cliente nao ve HTTP.
func traduzir(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return ErrNaoEncontrado
	case http.StatusForbidden:
		return ErrProibido
	}
	return fmt.Errorf("servico remoto: %s", resp.Status)
}

func (c *ClienteRemoto) Obter(ctx context.Context, id int) (Perfil, error) {
	resp, err := c.fazer(ctx, http.MethodGet, "/perfis/"+strconv.Itoa(id), nil)
	if err != nil {
		return Perfil{}, fmt.Errorf("chamando servico remoto: %w", err)
	}
	defer resp.Body.Close()
	if err := traduzir(resp); err != nil {
		return Perfil{}, err
	}
	var p Perfil
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return Perfil{}, fmt.Errorf("resposta ilegivel: %w", err)
	}
	return p, nil
}

func (c *ClienteRemoto) Atualizar(ctx context.Context, p Perfil) error {
	corpo, _ := json.Marshal(p)
	resp, err := c.fazer(ctx, http.MethodPut, "/perfis/"+strconv.Itoa(p.ID), corpo)
	if err != nil {
		return fmt.Errorf("chamando servico remoto: %w", err)
	}
	defer resp.Body.Close()
	return traduzir(resp)
}
