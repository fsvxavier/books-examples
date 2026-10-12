// Package repositoriotest traz o teste de contrato de aplicacao.Repositorio.
// Toda implementação (memória, PostgreSQL, a que vier) deve passar nele: é o
// que garante que uma pode substituir a outra sem que o caso de uso perceba.
package repositoriotest

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/dominio"
)

// Criado é o instante usado nos links do teste. O contrato exige só precisão
// de microssegundos para o campo Criado, porque é o que o PostgreSQL guarda.
var Criado = time.Date(2026, time.October, 5, 9, 0, 0, 123456000, time.UTC)

// Executar roda o contrato. novo deve devolver um repositório VAZIO a cada chamada.
func Executar(t *testing.T, novo func(t *testing.T) aplicacao.Repositorio) {
	t.Helper()

	link := func(codigo string) dominio.Link {
		l, err := dominio.NovoLink(codigo, "https://go.dev/doc/", Criado)
		if err != nil {
			t.Fatal(err)
		}
		return l
	}

	t.Run("salvar e buscar", func(t *testing.T) {
		r := novo(t)
		l := link("abc234")
		if err := r.Salvar(t.Context(), l); err != nil {
			t.Fatal(err)
		}
		got, err := r.Buscar(t.Context(), "abc234")
		if err != nil {
			t.Fatal(err)
		}
		if got.Codigo != l.Codigo || got.Destino != l.Destino || got.Acessos != 0 || !got.Criado.Equal(l.Criado) {
			t.Errorf("Buscar = %+v, esperado %+v", got, l)
		}
	})

	t.Run("código repetido", func(t *testing.T) {
		r := novo(t)
		if err := r.Salvar(t.Context(), link("abc234")); err != nil {
			t.Fatal(err)
		}
		if err := r.Salvar(t.Context(), link("abc234")); !errors.Is(err, dominio.ErrCodigoEmUso) {
			t.Fatalf("erro = %v, esperado ErrCodigoEmUso", err)
		}
	})

	t.Run("inexistente", func(t *testing.T) {
		r := novo(t)
		if _, err := r.Buscar(t.Context(), "zzzzzz"); !errors.Is(err, dominio.ErrNaoEncontrado) {
			t.Errorf("Buscar: %v", err)
		}
		if _, err := r.RegistrarAcesso(t.Context(), "zzzzzz"); !errors.Is(err, dominio.ErrNaoEncontrado) {
			t.Errorf("RegistrarAcesso: %v", err)
		}
	})

	t.Run("registrar acesso devolve o link atualizado", func(t *testing.T) {
		r := novo(t)
		_ = r.Salvar(t.Context(), link("abc234"))
		for i := 1; i <= 3; i++ {
			l, err := r.RegistrarAcesso(t.Context(), "abc234")
			if err != nil || l.Acessos != i {
				t.Fatalf("acesso %d: %+v, %v", i, l, err)
			}
		}
		if l, _ := r.Buscar(t.Context(), "abc234"); l.Acessos != 3 {
			t.Errorf("Buscar depois: acessos = %d", l.Acessos)
		}
	})

	t.Run("salvar concorrente do mesmo código tem um só vencedor", func(t *testing.T) {
		r := novo(t)
		var wg sync.WaitGroup
		var mu sync.Mutex
		venceram, perderam := 0, 0
		for range 20 {
			wg.Go(func() {
				err := r.Salvar(t.Context(), link("abc234"))
				mu.Lock()
				defer mu.Unlock()
				switch {
				case err == nil:
					venceram++
				case errors.Is(err, dominio.ErrCodigoEmUso):
					perderam++
				default:
					t.Errorf("erro inesperado: %v", err)
				}
			})
		}
		wg.Wait()
		if venceram != 1 || perderam != 19 {
			t.Errorf("venceram=%d perderam=%d, esperado 1 e 19", venceram, perderam)
		}
	})

	t.Run("acessos concorrentes não se perdem", func(t *testing.T) {
		r := novo(t)
		_ = r.Salvar(t.Context(), link("abc234"))
		var wg sync.WaitGroup
		for range 100 {
			wg.Go(func() { _, _ = r.RegistrarAcesso(t.Context(), "abc234") })
		}
		wg.Wait()
		if l, _ := r.Buscar(t.Context(), "abc234"); l.Acessos != 100 {
			t.Errorf("acessos = %d, esperado 100", l.Acessos)
		}
	})
}
