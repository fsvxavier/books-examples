package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metricas struct {
	Requisicoes *prometheus.CounterVec
	Duracao     *prometheus.HistogramVec
	EmVoo       prometheus.Gauge
}

func NovasMetricas(reg prometheus.Registerer) *Metricas {
	m := &Metricas{
		Requisicoes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requisicoes HTTP atendidas.",
		}, []string{"metodo", "rota", "status"}),
		Duracao: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duracao das requisicoes HTTP.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}, []string{"metodo", "rota"}),
		EmVoo: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Requisicoes sendo atendidas agora.",
		}),
	}
	reg.MustRegister(m.Requisicoes, m.Duracao, m.EmVoo)
	return m
}

type gravador struct {
	http.ResponseWriter
	status int
}

func (g *gravador) WriteHeader(c int)           { g.status = c; g.ResponseWriter.WriteHeader(c) }
func (g *gravador) Unwrap() http.ResponseWriter { return g.ResponseWriter }

// Instrumentar mede cada requisicao. A rota vem do PADRAO registrado no mux
// (r.Pattern, Go 1.23), nunca do caminho bruto: assim /produtos/1 e /produtos/2
// viram a mesma serie.
func (m *Metricas) Instrumentar(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.EmVoo.Inc()
		defer m.EmVoo.Dec()
		g := &gravador{ResponseWriter: w, status: 200}
		ini := time.Now()
		proximo.ServeHTTP(g, r)
		rota := r.Pattern
		if rota == "" {
			rota = "desconhecida"
		}
		m.Requisicoes.WithLabelValues(r.Method, rota, fmt.Sprint(g.status)).Inc()
		m.Duracao.WithLabelValues(r.Method, rota).Observe(time.Since(ini).Seconds())
	})
}

func main() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	m := NovasMetricas(reg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /produtos/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") == "0" {
			http.Error(w, "nao existe", http.StatusNotFound)
			return
		}
		time.Sleep(8 * time.Millisecond)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("POST /pedidos", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(30 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
	})
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	srv := httptest.NewServer(m.Instrumentar(mux))
	defer srv.Close()

	chamar := func(metodo, caminho string) {
		req, _ := http.NewRequest(metodo, srv.URL+caminho, nil)
		resp, err := srv.Client().Do(req)
		if err != nil {
			panic(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	for i := 1; i <= 5; i++ {
		chamar("GET", fmt.Sprintf("/produtos/%d", i))
	}
	chamar("GET", "/produtos/0")
	chamar("POST", "/pedidos")
	chamar("GET", "/nao-existe")

	resp, _ := srv.Client().Get(srv.URL + "/metrics")
	corpo, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	for _, linha := range strings.Split(string(corpo), "\n") {
		if strings.HasPrefix(linha, "http_") || strings.HasPrefix(linha, "# TYPE http_") {
			fmt.Println(linha)
		}
	}
}
