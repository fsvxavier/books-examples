package main

import (
	"fmt"
	"math"
)

func antigo(v float64) float64 {
	if v > 1000 {
		return v * 0.1
	}
	return v * 0.05
}

func main() {
	v := 70.0 / 100
	d := antigo(v)
	fmt.Printf("70 centavos como float64: %.20f\n", v)
	fmt.Printf("desconto antigo:          %.20f\n", d)
	fmt.Printf("desconto * 100:           %.20f -> Round = %v\n", d*100, math.Round(d*100))
	total, baixo, cima := 0, 0, 0
	for c := int64(0); c <= 300_000; c++ {
		pct := int64(5)
		if c > 100_000 {
			pct = 10
		}
		if c*pct%100 != 50 {
			continue
		}
		total++
		got := int64(math.Round(antigo(float64(c)/100) * 100))
		want := (c*pct + 50) / 100
		if got == want {
			cima++
		} else {
			baixo++
		}
	}
	fmt.Printf("casos de meio centavo exato: %d | antigo acertou (meio para cima): %d | antigo errou (caiu para baixo): %d\n", total, cima, baixo)
}
