package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// Interface do cliente.
type Verificador interface {
	Verificar(ctx context.Context, documento string) error
}

// VerificadorFunc adapta uma FUNCAO a interface, como http.HandlerFunc faz com http.Handler.
type VerificadorFunc func(ctx context.Context, documento string) error

func (f VerificadorFunc) Verificar(ctx context.Context, d string) error { return f(ctx, d) }

// Um tipo com estado tambem satisfaz a interface.
type ListaNegra struct{ bloqueados map[string]bool }

func (l ListaNegra) Verificar(_ context.Context, d string) error {
	if l.bloqueados[d] {
		return errors.New("documento na lista negra")
	}
	return nil
}

// Em (sequencia) combina verificadores; ele proprio e um Verificador.
func Todos(vs ...Verificador) Verificador {
	return VerificadorFunc(func(ctx context.Context, d string) error {
		for _, v := range vs {
			if err := v.Verificar(ctx, d); err != nil {
				return err
			}
		}
		return nil
	})
}

func main() {
	tamanho := VerificadorFunc(func(_ context.Context, d string) error {
		if len(d) != 11 {
			return fmt.Errorf("documento com %d digitos, esperava 11", len(d))
		}
		return nil
	})

	// Adaptando uma chamada HTTP existente (um endpoint) a mesma interface, por funcao.
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("doc") == "99999999999" {
			http.Error(w, "reprovado", http.StatusForbidden)
		}
	}))
	defer api.Close()
	remoto := VerificadorFunc(func(ctx context.Context, d string) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, api.URL+"?doc="+d, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("servico remoto recusou: %s", resp.Status)
		}
		return nil
	})

	v := Todos(tamanho, ListaNegra{map[string]bool{"11111111111": true}}, remoto)
	for _, d := range []string{"12345678901", "123", "11111111111", "99999999999"} {
		fmt.Printf("%-12s -> %v\n", d, v.Verificar(context.Background(), d))
	}
}
