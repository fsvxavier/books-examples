// Package sdk mostra um cliente de API IMUTAVEL: o Builder monta o cliente base e cada
// "Com..." devolve um CLONE modificado (Builder + Prototype juntos).
package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"time"
)

type Cliente struct {
	base       *url.URL
	cabecalhos map[string]string
	timeout    time.Duration
	http       *http.Client
}

var ErrConfig = errors.New("configuracao invalida")

// ---- Builder: montagem inicial, com validacao no Construir ----

type Builder struct {
	url        string
	cabecalhos map[string]string
	timeout    time.Duration
	erros      []error
}

func Novo(baseURL string) *Builder {
	return &Builder{url: baseURL, cabecalhos: map[string]string{"User-Agent": "sdk-livro/1.0"}, timeout: 10 * time.Second}
}

func (b *Builder) Token(t string) *Builder {
	if t == "" {
		b.erros = append(b.erros, errors.New("token vazio"))
		return b
	}
	b.cabecalhos["Authorization"] = "Bearer " + t
	return b
}

func (b *Builder) Timeout(d time.Duration) *Builder {
	if d <= 0 {
		b.erros = append(b.erros, fmt.Errorf("timeout %v deve ser positivo", d))
		return b
	}
	b.timeout = d
	return b
}

func (b *Builder) Construir() (*Cliente, error) {
	u, err := url.Parse(b.url)
	if err != nil || u.Scheme == "" || u.Host == "" {
		b.erros = append(b.erros, fmt.Errorf("url base invalida: %q", b.url))
	}
	if len(b.erros) > 0 {
		return nil, fmt.Errorf("%w: %w", ErrConfig, errors.Join(b.erros...))
	}
	return &Cliente{
		base:       u,
		cabecalhos: maps.Clone(b.cabecalhos),
		timeout:    b.timeout,
		http:       &http.Client{Timeout: b.timeout},
	}, nil
}

// ---- Prototype: clones profundos para derivar clientes ----

func (c *Cliente) clone() *Cliente {
	novo := *c
	novo.cabecalhos = maps.Clone(c.cabecalhos)
	u := *c.base
	novo.base = &u
	hc := *c.http
	novo.http = &hc
	return &novo
}

// ComCabecalho devolve um NOVO cliente; o original nao muda.
func (c *Cliente) ComCabecalho(k, v string) *Cliente {
	n := c.clone()
	n.cabecalhos[k] = v
	return n
}

func (c *Cliente) ComTimeout(d time.Duration) *Cliente {
	n := c.clone()
	n.timeout = d
	n.http.Timeout = d
	return n
}

func (c *Cliente) Timeout() time.Duration { return c.timeout }

// Get faz uma requisicao com os cabecalhos do cliente.
func (c *Cliente) Get(ctx context.Context, caminho string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.base.JoinPath(caminho).String(), nil)
	if err != nil {
		return "", err
	}
	for k, v := range c.cabecalhos {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return string(b), err
}
