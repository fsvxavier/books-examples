package main

import (
	"errors"
	"sync"
	"testing"
)

func TestMesmaChaveMesmoPonteiro(t *testing.T) {
	f := NovaFabrica(criarTipo)
	a, _ := f.Obter(Chave{"a", "x"})
	b, _ := f.Obter(Chave{"a", "x"})
	c, _ := f.Obter(Chave{"a", "y"})
	if a != b || a == c || f.Criados != 2 {
		t.Fatalf("a==b:%v a==c:%v criados:%d", a == b, a == c, f.Criados)
	}
}

func TestErroNaoEntraNoPool(t *testing.T) {
	f := NovaFabrica(criarTipo)
	for range 3 {
		if _, err := f.Obter(Chave{}); !errors.Is(err, ErrChave) {
			t.Fatalf("err = %v", err)
		}
	}
	if f.Tamanho() != 0 || f.Criados != 0 {
		t.Fatalf("pool=%d criados=%d", f.Tamanho(), f.Criados)
	}
}

// Mil goroutines pedindo as mesmas 5 chaves: exatamente 5 flyweights existem.
func TestConcorrenciaCriaUmaVez(t *testing.T) {
	f := NovaFabrica(criarTipo)
	chaves := []Chave{{"a", "1"}, {"b", "1"}, {"c", "1"}, {"d", "1"}, {"e", "1"}}
	vistos := make([]map[*Tipo]bool, len(chaves))
	var mu sync.Mutex
	for i := range vistos {
		vistos[i] = map[*Tipo]bool{}
	}
	var wg sync.WaitGroup
	for i := range 1000 {
		wg.Go(func() {
			k := i % len(chaves)
			v, err := f.Obter(chaves[k])
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			vistos[k][v] = true
			mu.Unlock()
		})
	}
	wg.Wait()
	if f.Criados != 5 {
		t.Fatalf("criados = %d, quero 5", f.Criados)
	}
	for k, m := range vistos {
		if len(m) != 1 {
			t.Fatalf("chave %d viu %d instancias distintas", k, len(m))
		}
	}
}

func TestImutabilidade(t *testing.T) {
	f := NovaFabrica(criarTipo)
	a, _ := f.Obter(Chave{"a", "x"})
	a.Textura()[0] = 7
	if a.Textura()[0] != 0 {
		t.Fatal("o estado interno foi alterado pela copia")
	}
}

func BenchmarkObterExistente(b *testing.B) {
	f := NovaFabrica(criarTipo)
	f.Obter(Chave{"a", "x"})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f.Obter(Chave{"a", "x"})
		}
	})
}
