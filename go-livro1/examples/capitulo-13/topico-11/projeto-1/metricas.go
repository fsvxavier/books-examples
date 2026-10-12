package metricas

import "time"

// Metricas acumula estatísticas de um servidor. NOTA: esta versão tem corridas de dados.
type Metricas struct {
	total   int            // contador de requisições
	porRota map[string]int // contagem por rota
	ultima  time.Time      // horário da última requisição
	ativo   bool           // indica se o coletor está ligado
}

func Nova() *Metricas {
	return &Metricas{porRota: make(map[string]int), ativo: true}
}

func (m *Metricas) Registrar(rota string) {
	if !m.ativo {
		return
	}
	m.total++
	m.porRota[rota]++
	m.ultima = time.Now()
}

func (m *Metricas) Total() int           { return m.total }
func (m *Metricas) Rota(rota string) int { return m.porRota[rota] }
func (m *Metricas) Ultima() time.Time    { return m.ultima }
func (m *Metricas) Desligar()            { m.ativo = false }
