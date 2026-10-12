package main

import (
	"math/rand/v2"
	"testing"
)

func gerar(r *rand.Rand, prof int) Expr {
	if prof == 0 || r.IntN(4) == 0 {
		if r.IntN(2) == 0 {
			return Num{float64(r.IntN(5))} // 0..4: exatos em float64
		}
		return Var{string(rune('x' + r.IntN(3)))}
	}
	switch r.IntN(3) {
	case 0:
		return Soma{gerar(r, prof-1), gerar(r, prof-1)}
	case 1:
		return Mul{gerar(r, prof-1), gerar(r, prof-1)}
	}
	return Neg{gerar(r, prof-1)}
}

func tamanho(e Expr) int {
	switch e := e.(type) {
	case Soma:
		return 1 + tamanho(e.E) + tamanho(e.D)
	case Mul:
		return 1 + tamanho(e.E) + tamanho(e.D)
	case Neg:
		return 1 + tamanho(e.E)
	}
	return 1
}

// Propriedades: simplificar preserva o valor e nunca aumenta a árvore;
// simplificar é idempotente.
func TestSimplificarPropriedades(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 22))
	env := Avaliar{map[string]float64{"x": 2, "y": 3, "z": 4}}
	menores := 0
	for i := range 2000 {
		e := gerar(r, 6)
		s := Aceitar[Expr](e, Simplificar{})
		if a, b := Aceitar[Resultado](e, env), Aceitar[Resultado](s, env); a.V != b.V {
			t.Fatalf("caso %d: valor mudou %v != %v em %s", i, a.V, b.V, Aceitar[string](e, Imprimir{}))
		}
		if tamanho(s) > tamanho(e) {
			t.Fatalf("caso %d: a árvore cresceu", i)
		}
		if tamanho(s) < tamanho(e) {
			menores++
		}
		if Aceitar[string](Aceitar[Expr](s, Simplificar{}), Imprimir{}) != Aceitar[string](s, Imprimir{}) {
			t.Fatalf("caso %d: não é idempotente", i)
		}
	}
	t.Logf("2000 expressões; %d ficaram menores", menores)
}
