package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime/metrics"
	"sync"
	"sync/atomic"
	"time"
)

// Verificacao e uma dependencia que o servico precisa para funcionar.
type Verificacao func(ctx context.Context) error

type Saude struct {
	mu      sync.Mutex
	checks  map[string]Verificacao
	pronto  atomic.Bool
	timeout time.Duration
}

func NovaSaude(timeout time.Duration) *Saude {
	return &Saude{checks: map[string]Verificacao{}, timeout: timeout}
}

func (s *Saude) Registrar(nome string, v Verificacao) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checks[nome] = v
}

// Viva responde sempre 200 enquanto o processo consegue atender: nao consulta dependencias.
func (s *Saude) Viva(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "vivo")
}

// Pronta consulta as dependencias, em paralelo, cada uma com prazo.
func (s *Saude) Pronta(w http.ResponseWriter, r *http.Request) {
	if !s.pronto.Load() {
		http.Error(w, "iniciando ou encerrando", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	s.mu.Lock()
	copia := make(map[string]Verificacao, len(s.checks))
	for k, v := range s.checks {
		copia[k] = v
	}
	s.mu.Unlock()

	var wg sync.WaitGroup
	var rmu sync.Mutex
	resultado := map[string]string{}
	falhou := false
	for nome, v := range copia {
		wg.Go(func() {
			err := v(ctx)
			rmu.Lock()
			defer rmu.Unlock()
			if err != nil {
				resultado[nome] = err.Error()
				falhou = true
			} else {
				resultado[nome] = "ok"
			}
		})
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	if falhou {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resultado)
}

func main() {
	s := NovaSaude(200 * time.Millisecond)
	var bancoOk atomic.Bool
	bancoOk.Store(true)
	s.Registrar("banco", func(ctx context.Context) error {
		if !bancoOk.Load() {
			return errors.New("conexao recusada")
		}
		return nil
	})
	s.Registrar("cache", func(ctx context.Context) error {
		select {
		case <-time.After(time.Second): // cache lento
			return nil
		case <-ctx.Done():
			return fmt.Errorf("cache: %w", ctx.Err())
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.Viva)
	mux.HandleFunc("GET /readyz", s.Pronta)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	get := func(rotulo, caminho string) {
		ini := time.Now()
		resp, _ := http.Get(srv.URL + caminho)
		var corpo [200]byte
		n, _ := resp.Body.Read(corpo[:])
		resp.Body.Close()
		fmt.Printf("%-24s %d em %-5s %s", rotulo, resp.StatusCode, time.Since(ini).Round(50*time.Millisecond), corpo[:n])
		if n > 0 && corpo[n-1] != '\n' {
			fmt.Println()
		}
	}
	get("viva (processo ok)", "/healthz")
	get("pronta antes de iniciar", "/readyz")
	s.pronto.Store(true)
	get("pronta (cache lento)", "/readyz")

	// runtime/metrics: leitura barata, sem parar o mundo
	amostras := []metrics.Sample{
		{Name: "/sched/goroutines:goroutines"},
		{Name: "/memory/classes/heap/objects:bytes"},
		{Name: "/gc/cycles/total:gc-cycles"},
	}
	metrics.Read(amostras)
	fmt.Println()
	for _, a := range amostras {
		fmt.Printf("%-40s %v\n", a.Name, a.Value.Uint64())
	}
}
