package armazem_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"exemplos/livro4/capitulo-2/topico-5/encurtador/internal/armazem"
)

func TestSalvarEBuscar(t *testing.T) {
	a := armazem.Novo()
	l := armazem.Link{Codigo: "abc234", Destino: "https://go.dev", Criado: time.Unix(0, 0)}
	if err := a.Salvar(l); err != nil {
		t.Fatal(err)
	}
	if err := a.Salvar(l); !errors.Is(err, armazem.ErrCodigoEmUso) {
		t.Fatalf("segundo Salvar: %v, esperado ErrCodigoEmUso", err)
	}
	got, err := a.Buscar("abc234")
	if err != nil || got != l {
		t.Fatalf("Buscar = %+v, %v", got, err)
	}
	if _, err := a.Buscar("zzzzzz"); !errors.Is(err, armazem.ErrNaoEncontrado) {
		t.Fatalf("Buscar inexistente: %v", err)
	}
}

func TestBuscarDevolveCopia(t *testing.T) {
	a := armazem.Novo()
	_ = a.Salvar(armazem.Link{Codigo: "abc234", Destino: "https://go.dev"})
	l, _ := a.Buscar("abc234")
	l.Acessos = 99 // alterar a cópia não pode afetar o armazém
	again, _ := a.Buscar("abc234")
	if again.Acessos != 0 {
		t.Fatalf("o armazém vazou o estado interno: %+v", again)
	}
}

func TestRegistrarAcessoConcorrente(t *testing.T) {
	a := armazem.Novo()
	_ = a.Salvar(armazem.Link{Codigo: "abc234", Destino: "https://go.dev"})
	var wg sync.WaitGroup
	for range 500 {
		wg.Go(func() { _, _ = a.RegistrarAcesso("abc234") })
	}
	wg.Wait()
	l, _ := a.Buscar("abc234")
	if l.Acessos != 500 {
		t.Fatalf("acessos = %d, esperado 500", l.Acessos)
	}
}
