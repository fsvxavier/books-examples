package metricas

import (
	"sync"
	"testing"
)

func TestRegistrarConcorrente(t *testing.T) {
	m := Nova()
	rotas := []string{"/a", "/b", "/c"}

	var wg sync.WaitGroup
	for g := range 8 {
		wg.Go(func() {
			for i := range 500 {
				m.Registrar(rotas[(g+i)%len(rotas)])
				_ = m.Total()
				_ = m.Ultima()
			}
		})
	}
	wg.Go(func() { m.Desligar() }) // outra goroutine desliga durante o teste
	wg.Wait()

	if m.Total() > 4000 {
		t.Errorf("total = %d, maior que o número de chamadas", m.Total())
	}
}
