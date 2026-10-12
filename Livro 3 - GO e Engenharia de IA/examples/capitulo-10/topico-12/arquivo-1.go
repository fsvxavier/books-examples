package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Registro próprio: nada de init() nem de registro global.
	reg := prometheus.NewRegistry()
	casos := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "avalia_casos_total",
		Help: "Casos julgados, por resultado.",
	}, []string{"resultado"})
	duracao := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "avalia_chamada_duracao_segundos",
		Help:    "Duração de cada chamada ao juiz.",
		Buckets: []float64{0.5, 1, 2, 5},
	})
	reg.MustRegister(casos, duracao)

	for _, d := range []float64{0.4, 0.7, 0.9, 1.5, 3.2} {
		duracao.Observe(d)
	}
	casos.WithLabelValues("aprovado").Add(3)
	casos.WithLabelValues("reprovado").Add(2)

	// O Handler serve o registro no formato de texto do Prometheus.
	rec := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	fmt.Print(rec.Body.String())
}
