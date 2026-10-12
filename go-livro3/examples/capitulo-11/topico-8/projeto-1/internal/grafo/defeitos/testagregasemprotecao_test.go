//go:build demo

package defeitos

import (
	"context"
	"sort"
	"sync"
	"testing"

	"exemplo.com/agentes/internal/grafo"
)

// Defeito clássico: três pesquisadores em paralelo gravando no mesmo mapa e
// na mesma fatia, sem proteção.
func TestAgregaSemProtecao(t *testing.T) {
	fatos := map[string]string{}
	var ordem []string
	temas := []string{"goroutines", "canais", "mutex"}
	_, err := grafo.Mapear(t.Context(), temas, 3, func(_ context.Context, tema string) (struct{}, error) {
		fatos[tema] = "fato sobre " + tema // escrita concorrente no mapa
		ordem = append(ordem, tema)        // escrita concorrente na fatia
		return struct{}{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fatos) != 3 {
		t.Errorf("fatos = %d, quer 3", len(fatos))
	}
}

// Correção 1: o Mapear já devolve os resultados na ordem das entradas, então
// não é preciso agregar nada por fora.
func TestAgregaPeloRetorno(t *testing.T) {
	temas := []string{"goroutines", "canais", "mutex"}
	fatos, err := grafo.Mapear(t.Context(), temas, 3, func(_ context.Context, tema string) (string, error) {
		return "fato sobre " + tema, nil
	})
	if err != nil || len(fatos) != 3 {
		t.Fatalf("fatos = %v, err = %v", fatos, err)
	}
}

// Correção 2: quando o estado compartilhado é inevitável, um mutex.
func TestAgregaComMutex(t *testing.T) {
	var (
		mu    sync.Mutex
		ordem []string
	)
	temas := []string{"goroutines", "canais", "mutex"}
	_, err := grafo.Mapear(t.Context(), temas, 3, func(_ context.Context, tema string) (struct{}, error) {
		mu.Lock()
		defer mu.Unlock()
		ordem = append(ordem, tema)
		return struct{}{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(ordem) // a ordem de chegada não é determinística
	if len(ordem) != 3 {
		t.Errorf("ordem = %v", ordem)
	}
}
