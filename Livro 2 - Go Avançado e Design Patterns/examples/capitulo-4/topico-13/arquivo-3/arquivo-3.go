package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	novo := func() prometheus.Counter {
		return prometheus.NewCounter(prometheus.CounterOpts{Name: "pedidos_total", Help: "pedidos"})
	}
	prometheus.MustRegister(novo())
	defer func() { fmt.Println("MustRegister duplicado:", recover()) }()
	prometheus.MustRegister(novo())
}
