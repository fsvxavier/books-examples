package nativo

import (
	"runtime"
	"testing"
	"time"
)

func TestCloseLibera(t *testing.T) {
	base := BlocosVivos()
	b, err := NovoBuffer(1024)
	if err != nil {
		t.Fatal(err)
	}
	copy(b.Bytes(), "ola")
	if BlocosVivos() != base+1 {
		t.Fatalf("vivos = %d, quer %d", BlocosVivos(), base+1)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != ErrFechado {
		t.Errorf("segundo Close = %v, quer ErrFechado", err)
	}
	if BlocosVivos() != base {
		t.Errorf("vivos = %d depois do Close, quer %d", BlocosVivos(), base)
	}
	if b.Bytes() != nil {
		t.Error("Bytes depois do Close deveria ser nil")
	}
}

func TestVisaoEnxergaAEscritaEmC(t *testing.T) {
	b, _ := NovoBuffer(8)
	defer b.Close()
	visao := b.Bytes()
	visao[3] = 42
	if got := b.Copia()[3]; got != 42 {
		t.Errorf("copia[3] = %d, quer 42", got)
	}
	copia := b.Copia()
	visao[3] = 7
	if copia[3] != 42 {
		t.Errorf("a copia mudou junto com a visao: %d", copia[3])
	}
}

func TestEsquecerOClose(t *testing.T) {
	base := BlocosVivos()
	func() {
		for range 10 {
			if _, err := NovoBuffer(4096); err != nil {
				t.Fatal(err)
			}
			// ninguem chama Close
		}
	}()
	t.Log("vivos logo depois de esquecer:", BlocosVivos()-base)
	for i := 0; i < 50 && BlocosVivos() > base; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
	if got := BlocosVivos() - base; got != 0 {
		t.Errorf("sobraram %d blocos", got)
	}
	t.Log("vivos depois do GC:", BlocosVivos()-base)
}
