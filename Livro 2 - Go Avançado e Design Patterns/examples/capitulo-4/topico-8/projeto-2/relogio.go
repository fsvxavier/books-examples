package relogio

import (
	"sync"
	"time"
)

// Relogio e o que o resto do sistema enxerga.
type Relogio interface{ Agora() time.Time }

type sistema struct{}

func (sistema) Agora() time.Time { return time.Now() }

// --- Singleton "classico": variavel de pacote + Once ---

var (
	once      sync.Once
	instancia Relogio
)

func Obter() Relogio {
	once.Do(func() { instancia = sistema{} })
	return instancia
}

// --- Variante testavel: o padrao pode ser trocado, e a troca e desfeita ---

var (
	mu     sync.RWMutex
	padrao Relogio = sistema{}
)

func Padrao() Relogio {
	mu.RLock()
	defer mu.RUnlock()
	return padrao
}

// Definir troca o relogio global e devolve a funcao que restaura o anterior.
func Definir(r Relogio) (restaurar func()) {
	mu.Lock()
	anterior := padrao
	padrao = r
	mu.Unlock()
	return func() {
		mu.Lock()
		padrao = anterior
		mu.Unlock()
	}
}
