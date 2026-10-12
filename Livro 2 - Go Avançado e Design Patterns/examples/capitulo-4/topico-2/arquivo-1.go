package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	ID       int64
	Ambiente string
}

var criadas atomic.Int64

func novaConfig() *Config {
	// simula um carregamento caro (leitura de arquivo, rede...)
	time.Sleep(2 * time.Millisecond)
	return &Config{ID: criadas.Add(1), Ambiente: "producao"}
}

var instancia *Config

// ObterIngenuo parece certo, mas tem corrida entre a checagem e a escrita.
func ObterIngenuo() *Config {
	if instancia == nil {
		instancia = novaConfig()
	}
	return instancia
}

func main() {
	const n = 50
	vistas := make([]*Config, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			vistas[i] = ObterIngenuo()
		})
	}
	wg.Wait()

	distintas := map[int64]bool{}
	for _, c := range vistas {
		distintas[c.ID] = true
	}
	fmt.Println("goroutines:", n)
	fmt.Println("configs criadas:", criadas.Load())
	fmt.Println("instancias distintas vistas pelos chamadores:", len(distintas))
}
