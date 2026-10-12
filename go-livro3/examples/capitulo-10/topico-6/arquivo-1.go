package main

import (
	"fmt"
	"math"
)

// kappa calcula o kappa de Cohen entre os vereditos de um juiz e a verdade.
func kappa(juiz, verdade []bool) float64 {
	var vp, fp, vn, fn float64
	for i := range juiz {
		switch {
		case juiz[i] && verdade[i]:
			vp++
		case juiz[i] && !verdade[i]:
			fp++
		case !juiz[i] && !verdade[i]:
			vn++
		default:
			fn++
		}
	}
	n := vp + fp + vn + fn
	po := (vp + vn) / n
	pJuizSim, pVerdadeSim := (vp+fp)/n, (vp+fn)/n
	pe := pJuizSim*pVerdadeSim + (1-pJuizSim)*(1-pVerdadeSim)
	if pe == 1 {
		return math.NaN()
	}
	return (po - pe) / (1 - pe)
}

// repete devolve um vetor com n cópias de v.
func repete(v bool, n int) []bool {
	s := make([]bool, n)
	for i := range s {
		s[i] = v
	}
	return s
}

func main() {
	// Verdade: 100 respostas, 30 fiéis e 70 infiéis.
	verdade := append(repete(true, 30), repete(false, 70)...)

	aprovaTudo := repete(true, 100)
	reprovaTudo := repete(false, 100)
	// Um juiz razoável: aprova 27 das 30 fiéis e reprova 55 das 70 infiéis.
	razoavel := append(append(repete(true, 27), repete(false, 3)...), append(repete(false, 55), repete(true, 15)...)...)

	for _, c := range []struct {
		nome   string
		juiz   []bool
		verdad []bool
	}{{"aprova tudo", aprovaTudo, verdade}, {"reprova tudo", reprovaTudo, verdade}, {"juiz razoável", razoavel, verdade}} {
		acertos := 0
		for i := range c.juiz {
			if c.juiz[i] == c.verdad[i] {
				acertos++
			}
		}
		fmt.Printf("%-14s acurácia %.2f   kappa %.3f\n", c.nome, float64(acertos)/100, kappa(c.juiz, c.verdad))
	}
}
