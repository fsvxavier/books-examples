package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestOrdemDeInscricao(t *testing.T) {
	b := NovoBarramento()
	var ordem []int
	for i := range 3 {
		b.Inscrever("t", func(any) { ordem = append(ordem, i) })
	}
	b.Publicar("t", nil)
	if len(ordem) != 3 || ordem[0] != 0 || ordem[2] != 2 {
		t.Fatalf("ordem = %v", ordem)
	}
}

func TestCancelarIdempotenteECirurgico(t *testing.T) {
	b := NovoBarramento()
	var a, c int
	cancelA := b.Inscrever("t", func(any) { a++ })
	b.Inscrever("t", func(any) { c++ })
	cancelA()
	cancelA()
	b.Publicar("t", nil)
	if a != 0 || c != 1 || b.Inscritos("t") != 1 {
		t.Fatalf("a=%d c=%d inscritos=%d", a, c, b.Inscritos("t"))
	}
}

func TestCancelarDentroDaEntregaNaoTrava(t *testing.T) {
	b := NovoBarramento()
	var cancelar func()
	cancelar = b.Inscrever("t", func(any) { cancelar() })
	if n, _ := b.Publicar("t", nil); n != 1 || b.Inscritos("t") != 0 {
		t.Fatalf("n=%d inscritos=%d", n, b.Inscritos("t"))
	}
}

func TestPanicIsolado(t *testing.T) {
	b := NovoBarramento()
	chamou := false
	b.Inscrever("t", func(any) { panic("x") })
	b.Inscrever("t", func(any) { chamou = true })
	_, err := b.Publicar("t", nil)
	if err == nil || !chamou {
		t.Fatalf("err=%v chamou=%v", err, chamou)
	}
}

func TestConcorrencia(t *testing.T) {
	b := NovoBarramento()
	var total atomic.Int64
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			for range 50 {
				cancelar := b.Inscrever("t", func(any) { total.Add(1) })
				b.Publicar("t", nil)
				cancelar()
			}
		})
	}
	wg.Wait()
	if b.Inscritos("t") != 0 || total.Load() < 1000 {
		t.Fatalf("inscritos=%d total=%d", b.Inscritos("t"), total.Load())
	}
}
