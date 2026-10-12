// services/frete.go
package services

import "exemplo.com/ciclo/models"

func CalcularFrete(valor int64) int64 {
	if valor > 20000 {
		return 0
	}
	return 1290
}

func ResumoDoPedido(p models.Pedido) string { return "pedido pronto" }
