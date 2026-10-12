package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// consulta simula um trabalho de 300 ms que respeita o contexto.
func consulta(ctx context.Context, nome string) string {
	select {
	case <-time.After(300 * time.Millisecond):
		return nome + ": terminou o trabalho completo"
	case <-ctx.Done():
		return nome + ": interrompido (" + ctx.Err().Error() + ")"
	}
}

func main() {
	var mu sync.Mutex
	var log []string
	registrar := func(s string) {
		mu.Lock()
		log = append(log, s)
		mu.Unlock()
	}
	var wg sync.WaitGroup

	mux := http.NewServeMux()
	// RUIM: Background solto, o cliente desiste e o trabalho segue.
	mux.HandleFunc("/ruim", func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()
		registrar(consulta(context.Background(), "/ruim"))
	})
	// BOM: propaga o contexto da requisição.
	mux.HandleFunc("/bom", func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()
		registrar(consulta(r.Context(), "/bom"))
	})
	// Armadilha: goroutine que sobrevive ao handler com r.Context().
	mux.HandleFunc("/armadilha", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		wg.Add(1)
		go func() {
			defer wg.Done()
			registrar(consulta(ctx, "/armadilha"))
		}()
		w.WriteHeader(http.StatusAccepted) // responde já; o trabalho deveria continuar em segundo plano
	})
	// Correto para trabalho que deve sobreviver à resposta.
	mux.HandleFunc("/desacoplado", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), time.Second)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()
			registrar(consulta(ctx, "/desacoplado"))
		}()
		w.WriteHeader(http.StatusAccepted)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// O cliente desiste depois de 50 ms em todas as rotas.
	for _, rota := range []string{"/ruim", "/bom", "/armadilha", "/desacoplado"} {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL+rota, nil)
		inicio := time.Now()
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			fmt.Printf("cliente %-13s resposta %d em %v\n", rota, resp.StatusCode, time.Since(inicio).Round(10*time.Millisecond))
		} else {
			fmt.Printf("cliente %-13s desistiu após %v\n", rota, time.Since(inicio).Round(10*time.Millisecond))
		}
		cancel()
	}
	wg.Wait()
	fmt.Println("o que o servidor registrou:")
	mu.Lock()
	for _, l := range log {
		fmt.Println("  ", l)
	}
	mu.Unlock()
}
