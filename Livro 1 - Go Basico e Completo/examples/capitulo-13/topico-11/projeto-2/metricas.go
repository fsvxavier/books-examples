package metricas

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metricas acumula estatísticas de um servidor. Seguro para uso concorrente.
type Metricas struct {
	total atomic.Int64 // um valor só, independente dos demais: atômico basta
	ativo atomic.Bool  // indicador simples: atômico basta

	mu      sync.Mutex // protege porRota e ultima, que mudam JUNTOS
	porRota map[string]int
	ultima  time.Time
}

func Nova() *Metricas {
	m := &Metricas{porRota: make(map[string]int)}
	m.ativo.Store(true)
	return m
}

func (m *Metricas) Registrar(rota string) {
	if !m.ativo.Load() {
		return
	}
	m.total.Add(1)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.porRota[rota]++
	m.ultima = time.Now()
}

func (m *Metricas) Total() int { return int(m.total.Load()) }

func (m *Metricas) Rota(rota string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.porRota[rota]
}

func (m *Metricas) Ultima() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ultima
}

func (m *Metricas) Desligar() { m.ativo.Store(false) }
