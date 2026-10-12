package main

import "fmt"

// memoria estima o estado de treino de um modelo hipotético com formato de 7B.
func memoria(args []string) error {
	const (
		d       = 4096
		camadas = 32
		total   = 6.7e9 // parâmetros da base (aproximado)
	)
	gib := func(b float64) float64 { return b / (1 << 30) }
	// Ajuste fino completo, precisão mista com Adam: 2 (pesos bf16) + 2 (gradiente bf16)
	// + 4 (cópia mestre fp32) + 4 (m) + 4 (v) = 16 bytes por parâmetro.
	full := total * 16
	fmt.Printf("modelo hipotético: %.1f bi de parâmetros, d=%d, %d camadas\n", total/1e9, d, camadas)
	fmt.Printf("ajuste fino completo (bf16 + Adam fp32): %.0f GiB de estado (sem ativações)\n", gib(full))
	fmt.Printf("checkpoint completo (cópia mestre + m + v = 12 B/param): %.0f GiB; só pesos bf16: %.0f GiB\n\n", gib(total*12), gib(total*2))
	fmt.Printf("%-26s %12s %14s %14s %14s\n", "LoRA em q,k,v,o (4 matrizes)", "param. LoRA", "estado total", "checkpoint", "x menor")
	for _, r := range []int{4, 8, 16, 64} {
		porMatriz := float64(r * (d + d))
		pl := porMatriz * 4 * camadas
		// Base congelada em bf16 (2 B) + adaptador com estado completo de treino (16 B).
		estado := total*2 + pl*16
		ck := pl * 12 // mestre fp32 + m + v do adaptador (a base não muda)
		fmt.Printf("r=%-24d %9.1f M %11.1f GiB %11.1f MiB %13.0fx\n", r, pl/1e6, gib(estado), ck/(1<<20), (total*12)/ck)
	}
	fmt.Println()
	fmt.Printf("QLoRA r=16 (base em 4 bits, ~0.5 B/param): estado ~ %.1f GiB\n", gib(total*0.5+16.8e6*16))
	return nil
}
