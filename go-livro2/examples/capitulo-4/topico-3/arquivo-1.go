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
	time.Sleep(2 * time.Millisecond)
	return &Config{ID: criadas.Add(1), Ambiente: "producao"}
}

var (
	mu        sync.Mutex
	instancia *Config
)

// ObterComMutex: correto, mas todo acesso paga a trava.
func ObterComMutex() *Config {
	mu.Lock()
	defer mu.Unlock()
	if instancia == nil {
		instancia = novaConfig()
	}
	return instancia
}

var (
	once    sync.Once
	oConfig *Config
)

// ObterComOnce: a forma classica.
func ObterComOnce() *Config {
	once.Do(func() { oConfig = novaConfig() })
	return oConfig
}

// ObterComOnceValue: a forma moderna (Go 1.21).
var obterOV = sync.OnceValue(novaConfig)

func ObterComOnceValue() *Config { return obterOV() }

func contar(nome string, obter func() *Config) {
	criadas.Store(0)
	const n = 50
	vistas := make([]*Config, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() { vistas[i] = obter() })
	}
	wg.Wait()
	d := map[int64]bool{}
	for _, c := range vistas {
		d[c.ID] = true
	}
	fmt.Printf("%-14s criadas=%d distintas=%d\n", nome, criadas.Load(), len(d))
}

func main() {
	contar("Mutex", ObterComMutex)
	contar("Once", ObterComOnce)
	contar("OnceValue", ObterComOnceValue)
}
