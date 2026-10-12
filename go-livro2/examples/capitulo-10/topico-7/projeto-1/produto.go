package main

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"strconv"
)

type Produto struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

type pagina struct {
	Itens   []Produto `json:"itens"`
	Proxima string    `json:"proxima"` // vazio na ultima pagina
}

// Cliente le uma API paginada.
type Cliente struct {
	Base string
	HTTP *http.Client
}

// Produtos devolve um iterador PREGUICOSO sobre TODOS os produtos, atravessando as paginas
// sem que o chamador saiba que existem paginas. Cada elemento vem com um erro (Seq2):
// o padrao para fontes que podem falhar no meio do caminho.
//
// E um iterador de USO UNICO por chamada: cada range faz novas requisicoes.
func (c *Cliente) Produtos(ctx context.Context) iter.Seq2[Produto, error] {
	return func(yield func(Produto, error) bool) {
		url := c.Base + "/produtos?pagina=1"
		for url != "" {
			pg, err := c.buscar(ctx, url)
			if err != nil {
				yield(Produto{}, err) // entrega o erro e ENCERRA: nao ha como continuar
				return
			}
			for _, p := range pg.Itens {
				if !yield(p, nil) {
					return // o chamador deu break: nao busca mais paginas
				}
			}
			url = pg.Proxima
		}
	}
}

func (c *Cliente) buscar(ctx context.Context, url string) (pagina, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return pagina{}, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return pagina{}, fmt.Errorf("buscando %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return pagina{}, fmt.Errorf("buscando %s: status %s", url, resp.Status)
	}
	var pg pagina
	if err := json.NewDecoder(resp.Body).Decode(&pg); err != nil {
		return pagina{}, fmt.Errorf("lendo %s: %w", url, err)
	}
	return pg, nil
}

// ---------- Servidor falso: 25 produtos, 10 por pagina ----------

type Servidor struct {
	Requisicoes    int
	FalharNaPagina int // 0 = nunca falha
}

func (s *Servidor) Handler(base func() string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Requisicoes++
		n, _ := strconv.Atoi(r.URL.Query().Get("pagina"))
		if s.FalharNaPagina != 0 && n == s.FalharNaPagina {
			http.Error(w, "indisponivel", http.StatusServiceUnavailable)
			return
		}
		const total, porPagina = 25, 10
		pg := pagina{}
		for id := (n-1)*porPagina + 1; id <= min(n*porPagina, total); id++ {
			pg.Itens = append(pg.Itens, Produto{id, "produto-" + strconv.Itoa(id)})
		}
		if n*porPagina < total {
			pg.Proxima = base() + "/produtos?pagina=" + strconv.Itoa(n+1)
		}
		json.NewEncoder(w).Encode(pg)
	})
}
