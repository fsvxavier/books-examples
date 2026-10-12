package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func servidor(t *testing.T) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s|%s|%s", r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("X-Tenant"))
	}))
	t.Cleanup(s.Close)
	return s
}

func TestBuilderValida(t *testing.T) {
	_, err := Novo("://ruim").Token("").Timeout(-1).Construir()
	if !errors.Is(err, ErrConfig) {
		t.Fatalf("esperava ErrConfig, veio %v", err)
	}
	for _, trecho := range []string{"token vazio", "timeout", "url base invalida"} {
		if !strings.Contains(err.Error(), trecho) {
			t.Errorf("faltou %q em: %v", trecho, err)
		}
	}
}

func TestDerivarNaoAlteraOOriginal(t *testing.T) {
	s := servidor(t)
	base, err := Novo(s.URL).Token("abc").Construir()
	if err != nil {
		t.Fatal(err)
	}
	tenantA := base.ComCabecalho("X-Tenant", "a")
	lento := tenantA.ComTimeout(3 * time.Second)

	for nome, c := range map[string]*Cliente{"base": base, "tenantA": tenantA, "lento": lento} {
		got, err := c.Get(t.Context(), "/p")
		if err != nil {
			t.Fatalf("%s: %v", nome, err)
		}
		t.Logf("%-8s -> %s (timeout %s)", nome, got, c.Timeout())
	}
	if base.Timeout() != 10*time.Second || tenantA.Timeout() != 10*time.Second || lento.Timeout() != 3*time.Second {
		t.Fatal("timeouts derivados vazaram")
	}
	if got, _ := base.Get(t.Context(), "/p"); got != "/p|Bearer abc|" {
		t.Fatalf("o cliente base ganhou o cabecalho do derivado: %q", got)
	}
}

// Clientes imutaveis podem ser usados por muitas goroutines SEM trava.
func TestUsoConcorrente(t *testing.T) {
	s := servidor(t)
	base, _ := Novo(s.URL).Token("abc").Construir()
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			c := base.ComCabecalho("X-Tenant", fmt.Sprint("t", i))
			got, err := c.Get(t.Context(), "/c")
			if err != nil || got != fmt.Sprintf("/c|Bearer abc|t%d", i) {
				t.Errorf("i=%d got=%q err=%v", i, got, err)
			}
		})
	}
	wg.Wait()
}
