package main

import (
	"math"
	"testing"
)

func TestCalcular(t *testing.T) {
	casos := []struct {
		nome   string
		vendas []float64
		quer   Estatisticas
	}{
		{"lista vazia", nil, Estatisticas{}},
		{"um valor", []float64{50}, Estatisticas{Total: 50, Media: 50, Maior: 50}},
		{"varios valores", []float64{10, 20, 30}, Estatisticas{Total: 60, Media: 20, Maior: 30}},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got := Calcular(c.vendas)
			if math.Abs(got.Total-c.quer.Total) > 1e-9 ||
				math.Abs(got.Media-c.quer.Media) > 1e-9 ||
				math.Abs(got.Maior-c.quer.Maior) > 1e-9 {
				t.Errorf("Calcular(%v) = %+v, quer %+v", c.vendas, got, c.quer)
			}
		})
	}
}
