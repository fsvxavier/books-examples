package main

import (
	"errors"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
)

func TestRestaurarNotificaOsObservadores(t *testing.T) {
	o := &Orcamento{}
	var vistos []Mudanca
	o.Observar(func(m Mudanca) { vistos = append(vistos, m) })

	_ = o.Adicionar("a", 100)
	foto := o.Salvar()
	_ = o.Adicionar("b", 200)
	o.Restaurar(foto)

	if len(vistos) != 3 {
		t.Fatalf("eventos = %d, quero 3", len(vistos))
	}
	if ult := vistos[2]; ult.Tipo != Restaurado || ult.Total != 100 || ult.Versao != 3 {
		t.Fatalf("último evento = %+v", ult)
	}
}

func TestLoteAtomico(t *testing.T) {
	o := &Orcamento{}
	eventos := 0
	o.Observar(func(Mudanca) { eventos++ })

	falha := errors.New("falhou")
	err := o.Lote(
		func(it *[]Item) error { *it = append(*it, Item{"a", 1}); return nil },
		func(*[]Item) error { return falha },
	)
	if !errors.Is(err, falha) {
		t.Fatalf("err = %v", err)
	}
	if o.Total() != 0 || eventos != 0 {
		t.Fatalf("total=%d eventos=%d: o lote falho deixou rastros", o.Total(), eventos)
	}

	_ = o.Lote(
		func(it *[]Item) error { *it = append(*it, Item{"a", 1}); return nil },
		func(it *[]Item) error { *it = append(*it, Item{"b", 2}); return nil },
	)
	if o.Total() != 3 || eventos != 1 {
		t.Fatalf("total=%d eventos=%d", o.Total(), eventos)
	}
}

func TestFotoNaoCompartilhaMemoria(t *testing.T) {
	o := &Orcamento{}
	_ = o.Adicionar("a", 100)
	foto := o.Salvar()
	_ = o.Remover("a")
	_ = o.Adicionar("b", 5)
	o.Restaurar(foto)
	if o.Total() != 100 {
		t.Fatalf("total = %d, quero 100", o.Total())
	}
}

func TestSessaoDesfazerRefazerPropriedade(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 5))
	for rodada := range 300 {
		s := NovaSessao(&Orcamento{})
		var totais []int // total ANTES de cada edição bem-sucedida
		for range 1 + r.IntN(20) {
			totais = append(totais, s.Total())
			if s.Total() > 0 && r.IntN(4) == 0 {
				if err := s.Remover("x"); err != nil {
					t.Fatal(err)
				}
			} else {
				_ = s.Adicionar("x", 1+r.IntN(1000))
			}
		}
		final := s.Total()

		var vistos []int
		for s.Desfazer() {
			vistos = append(vistos, s.Total())
		}
		var esperado []int
		for i := len(totais) - 1; i >= 0; i-- {
			esperado = append(esperado, totais[i])
		}
		if !slices.Equal(vistos, esperado) {
			t.Fatalf("rodada %d: desfazer = %v, esperado %v", rodada, vistos, esperado)
		}
		for s.Refazer() {
		}
		if s.Total() != final {
			t.Fatalf("rodada %d: refazer tudo = %d, esperado %d", rodada, s.Total(), final)
		}
	}
}

func TestRestaurarSobrescreveMudancasPosteriores(t *testing.T) {
	// Documenta o comportamento: restaurar é uma escrita que vale por
	// inteiro; qualquer mudança feita depois da foto (por quem for) se perde.
	o := &Orcamento{}
	foto := o.Salvar()
	_ = o.Adicionar("de outra goroutine", 500)
	o.Restaurar(foto)
	if o.Total() != 0 {
		t.Fatalf("total = %d", o.Total())
	}
}

func TestConcorrencia(t *testing.T) {
	o := &Orcamento{}
	painel := &Painel{}
	var mu sync.Mutex
	eventos := 0
	o.Observar(painel.Atualizar)
	o.Observar(func(Mudanca) { mu.Lock(); eventos++; mu.Unlock() })

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				_ = o.Adicionar("x", 1)
			}
		}()
	}
	wg.Wait()

	if eventos != 1600 {
		t.Fatalf("eventos = %d, quero 1600", eventos)
	}
	if painel.Total() != 1600 || o.Total() != 1600 {
		t.Fatalf("painel=%d real=%d", painel.Total(), o.Total())
	}
}
