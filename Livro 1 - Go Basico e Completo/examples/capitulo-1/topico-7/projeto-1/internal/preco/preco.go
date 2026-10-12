// Package preco concentra as regras de cálculo de preço da loja.
package preco

import "fmt"

// ComDesconto aplica um desconto percentual ao valor original.
// O percentual deve estar entre 0 e 100.
func ComDesconto(valor, percentual float64) (float64, error) {
	if valor < 0 {
		return 0, fmt.Errorf("valor negativo: %.2f", valor)
	}
	if percentual < 0 || percentual > 100 {
		return 0, fmt.Errorf("percentual fora do intervalo 0-100: %.2f", percentual)
	}
	return valor * (1 - percentual/100), nil
}
