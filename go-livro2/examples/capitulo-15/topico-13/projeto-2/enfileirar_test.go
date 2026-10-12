package ruim

import (
	"runtime"
	"testing"
	"time"
)

func TestEnviaTudo(t *testing.T) {
	Iniciar()
	for range 20 {
		Enfileirar("olá")
	}
	time.Sleep(400 * time.Millisecond)
	if Enviados() != 20 {
		t.Fatalf("esperava 20 envios, houve %d", Enviados())
	}
}

func TestNaoVazaGoroutines(t *testing.T) {
	antes := runtime.NumGoroutine()
	Iniciar()
	time.Sleep(50 * time.Millisecond)
	// não há como pedir para parar: o teste só pode constatar o vazamento
	if depois := runtime.NumGoroutine(); depois > antes {
		t.Fatalf("goroutines: %d -> %d, e nenhuma forma de encerrar", antes, depois)
	}
}
