package main

import "fmt"

// Estatisticas reúne os números que o relatório de vendas precisa.
type Estatisticas struct {
	Total float64
	Media float64
	Maior float64
}

// Calcular percorre as vendas uma única vez e devolve total, média e maior valor.
func Calcular(vendas []float64) Estatisticas {
	if len(vendas) == 0 {
		return Estatisticas{}
	}

	est := Estatisticas{Maior: vendas[0]}
	for _, v := range vendas {
		est.Total += v
		if v > est.Maior {
			est.Maior = v
		}
	}
	est.Media = est.Total / float64(len(vendas))
	return est
}

func main() {
	vendas := []float64{120.50, 89.90, 340.00, 75.30, 210.00}
	est := Calcular(vendas)

	fmt.Printf("Total vendido: R$ %.2f\n", est.Total)
	fmt.Printf("Média por venda: R$ %.2f\n", est.Media)
	fmt.Printf("Maior venda: R$ %.2f\n", est.Maior)
}
