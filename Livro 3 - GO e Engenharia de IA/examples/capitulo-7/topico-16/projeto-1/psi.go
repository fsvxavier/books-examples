// Package deriva mede o quanto a distribuição de uma feature em produção se
// afastou da distribuição vista no treino, com o PSI (Population Stability
// Index).
package deriva

import (
	"math"
	"slices"
	"sort"
)

// PSI compara `atual` com `esperado` usando faixas definidas pelos decis de
// `esperado`. Regra prática comum: abaixo de 0,1 estável; de 0,1 a 0,25
// atenção; acima de 0,25 mudou de verdade.
func PSI(esperado, atual []float64, faixas int) float64 {
	ord := slices.Clone(esperado)
	slices.Sort(ord)
	cortes := make([]float64, 0, faixas-1)
	for i := 1; i < faixas; i++ {
		cortes = append(cortes, ord[len(ord)*i/faixas])
	}
	cortes = slices.Compact(cortes) // feature quase constante gera cortes repetidos

	proporcoes := func(xs []float64) []float64 {
		cont := make([]float64, len(cortes)+1)
		for _, x := range xs {
			cont[sort.SearchFloat64s(cortes, x)]++
		}
		for i := range cont {
			cont[i] = math.Max(cont[i]/float64(len(xs)), 1e-4) // evita log(0)
		}
		return cont
	}
	e, a := proporcoes(esperado), proporcoes(atual)
	var psi float64
	for i := range e {
		psi += (a[i] - e[i]) * math.Log(a[i]/e[i])
	}
	return psi
}
